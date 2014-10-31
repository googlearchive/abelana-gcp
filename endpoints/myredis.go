// Copyright 2014 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package abelana

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
)

var server string

var (
	pool          *Pool
	redisPassword = ""
)

// Note - I looked at adapting both Gary Burd's pool system and Vites pools to AppEngine, but ran
// out of time as there were too many dependencies.

func redisInit() {
	server = redisExt
	pool = newPool(server, redisPassword)
}

func newPool(server, password string) *Pool {
	return &Pool{
		MaxIdle:     3,
		IdleTimeout: 115 * time.Second,
		Dial: func(cx appengine.Context) (Conn, error) {
			c, err := Dial(cx, "tcp", server)
			if err != nil {
				return nil, err
			}
			// if _, err := c.Do("AUTH", password); err != nil {
			// 	c.Close()
			// 	return nil, err
			// }
			return c, err
		},
		TestOnBorrow: func(c Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// iNowFollow is Called when we need to add a folower
func iNowFollow(cx appengine.Context, userID, followerID string) {

}

// addPhoto is called to add a photo. This is allways called from a Delay
func addPhoto(cx appengine.Context, photoID string) {
	err := func() error {
		s := strings.Split(photoID, ".")

		// s[0] = userid, s[1] = random photo
		userID := s[0]
		u, err := findUser(cx, userID)
		if err != nil {
			return fmt.Errorf("unable to find user %v %v", userID, err)
		}

		p := &Photo{photoID, time.Now().UTC().Unix()}

		conn := pool.Get(cx)
		defer conn.Close()

		set, err := Int(conn.Do("HSETNX", "IM:"+photoID, "date", p.Date)) // Set Date
		if (err != nil && err != ErrNil) || set == 0 {
			return fmt.Errorf("duplicate %v %v", err, set)
		}

		// TODO: Consider if these should be done in batches of 100 or so.

		// Add to each follower's list
		for _, personID := range u.People {
			conn.Send("LPUSH", "TL:"+personID, photoID)
		}
		conn.Flush()

		// Check the result and adjust list if nescessary.
		for _, personID := range u.People {
			v, err := Int(conn.Receive())
			if err != nil && err != ErrNil {
				cx.Errorf("Addphoto: TL:%v %v", personID, err)
			} else {
				if v > 2000 {
					_, err := conn.Do("RPOP", "TL:"+personID)
					if err != nil {
						cx.Errorf("AddPhoto: RPOP TL:%v %v", personID, err)
					}
				}
			}
		}
		k1 := datastore.NewKey(cx, "User", userID, 0, nil)
		k2 := datastore.NewKey(cx, "Photo", photoID, 0, k1)
		_, err = datastore.Put(cx, k2, p)
		if err != nil {
			return fmt.Errorf("datastore %v", err)
		}

		return nil
	}()
	if err != nil {
		cx.Errorf("addPhoto: %v", err)
	}
}

// getTimeline returns the user's Timeline, you could insert additional things here as well.
func getTimeline(cx appengine.Context, userID, lastid string) ([]TLEntry, error) {
	timeline := []TLEntry{}

	conn := pool.Get(cx)
	defer conn.Close()

	list, err := Strings(conn.Do("LRANGE", "TL:"+userID, 0, -1))
	if err != nil && err != ErrNil {
		cx.Errorf("GetTimeLine %v", err)
	}
	ix := 0

	if lastid != "0" {
		for i, item := range list {
			if item == lastid {
				ix = i
				break
			}
		}
	}
	timeline = make([]TLEntry, 0, timelineBatchSize)
	for i := 0; i < timelineBatchSize && i+ix < len(list); i++ {
		photoID := list[ix+i]

		v, err := Strings(conn.Do("HMGET", "IM:"+photoID, "date", userID, "flag"))
		if err != nil && err != ErrNil {
			cx.Errorf("GetTimeLine HMGET %v", err)
		}
		if v[2] != "" {
			flags, err := strconv.Atoi(v[2])
			if err == nil && flags > 1 {
				continue // skip flag'd images
			}
		}
		likes, err := Int(conn.Do("HLEN", "IM:"+photoID))
		if err != nil && err != ErrNil {
			cx.Errorf("GetTimeLine HLEN %v", err)
		}
		s := strings.Split(photoID, ".")
		dn, err := String(conn.Do("HGET", "HT:"+s[0], "dn"))
		if err != nil && err != ErrNil {
			cx.Errorf("GetTimeLine HLEN %v", err)
		}
		dt, err := strconv.ParseInt(v[0], 10, 64)
		te := TLEntry{dt, s[0], dn, photoID, likes - 1, v[1] == "1"}
		timeline = append(timeline, te)
	}
	return timeline, nil
}

// addUser adds the user to redis
func addUser(cx appengine.Context, userID, displayName string) {

	conn := pool.Get(cx)
	defer conn.Close()

	// See if we have done this already, block others.
	_, err := conn.Do("HSET", "HT:"+userID, "dn", displayName)
	if err != nil {
		cx.Errorf("addUser Exists %v", err)
	}
}

// getPersons finds all the display names
func getPersons(cx appengine.Context, personids []string) []Person {
	pl := []Person{}

	conn := pool.Get(cx)
	defer conn.Close()

	for _, personID := range personids {
		conn.Send("HGET", "HT:"+personID, "dn")
	}
	conn.Flush()
	for _, personID := range personids {
		dn, _ := String(conn.Receive())
		p := Person{"abelana#follower", personID, "", dn}
		pl = append(pl, p)
	}
	return pl
}

// like the user on redis
func like(cx appengine.Context, userID, photoID string) {

	conn := pool.Get(cx)
	defer conn.Close()

	_, err := Int(conn.Do("HSET", "IM:"+photoID, userID, "1"))
	if err != nil && err != ErrNil {
		cx.Errorf("like %v", err)
	}
}

// unlike
func unlike(cx appengine.Context, userID, photoID string) {
	conn := pool.Get(cx)
	defer conn.Close()

	_, err := Int(conn.Do("HDEL", "IM:"+photoID, userID))
	if err != nil && err != ErrNil {
		cx.Errorf("unlike %v", err)
	}
}

// flag will tell us that things may not be quite right with this image.
func flag(cx appengine.Context, userID, photoID string) {
	conn := pool.Get(cx)
	defer conn.Close()

	_, err := Int(conn.Do("HINCRBY", "IM:"+photoID, "flag", 1))
	if err != nil && err != ErrNil {
		cx.Errorf("unlike %v", err)
	}
}
