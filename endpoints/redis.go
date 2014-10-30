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
	"appengine/socket"

	"github.com/garyburd/redigo/redis"
)

var server string

// Note - I looked at adapting both Gary Burd's pool system and Vites pools to AppEngine, but ran
// out of time as there were too many dependencies.

func redisInit() {
	if appengine.IsDevAppServer() {
		server = redisExt
	} else {
		server = redisExt
	}
}

// iNowFollow is Called when we need to add a folower
func iNowFollow(cx appengine.Context, userID, followerID string) {

}

// addPhoto is called to add a photo.
func addPhoto(c appengine.Context, filename string) {
	err := func() error {
		s := strings.Split(filename, ".")
		if len(s) != 3 {
			return fmt.Errorf("bad filename format %v", filename)
		}
		// s[2] is the file extension
		userID, photoID, superID := s[0], s[1], s[0]+"."+s[1]
		u, err := findUser(c, userID)
		if err != nil {
			return fmt.Errorf("unable to find user %v %v", userID, err)
		}

		p := &Photo{photoID, time.Now().UTC().Unix()}

		socket, err := socket.Dial(c, "tcp", server)
		if err != nil {
			return fmt.Errorf("Dial %v", err)
		}
		conn := redis.NewConn(socket, 0, 0) // TODO 0 TO's for now
		defer conn.Close()

		// Set Date
		set, err := redis.Int(conn.Do("HSETNX", "IM:"+superID, "date", p.Date))
		if (err != nil && err != redis.ErrNil) || set == 0 {
			// block duplicate requests
			return fmt.Errorf("duplicate %v %v", err, set)
		}

		// TODO: Consider that at some point these should be done in batches of 100 or so.

		// Add to each follower's list
		for _, personID := range u.Persons {
			conn.Send("LPUSH", "TL:"+personID, photoID)
		}
		conn.Flush()

		// Check the result and adjust list if nescessary.
		for _, personID := range u.Persons {
			v, err := redis.Int(conn.Receive())
			if err != nil && err != redis.ErrNil {
				c.Errorf("AddPhoto: TL:%v %v", personID, err)
				continue
			}
			if v > 2000 {
				_, err := conn.Do("RPOP", "TL:"+personID)
				if err != nil {
					c.Errorf("AddPhoto RPOP %v", err)
				}
			}
		}
		k1 := datastore.NewKey(c, "User", s[0], 0, nil)
		k2 := datastore.NewKey(c, "Photo", s[1], 0, k1)
		_, err = datastore.Put(c, k2, p)
		if err != nil {
			return fmt.Errorf("datastore %v", err)
		}
		return nil
	}()
	if err != nil {
		c.Errorf("AddPhoto: %v", err)
	}
}

// getTimeline
func getTimeline(cx appengine.Context, userID, lastid string) ([]TLEntry, error) {
	var item string
	timeline := []TLEntry{}

	hc, err := socket.Dial(cx, "tcp", server)
	if err != nil {
		cx.Errorf("GetTimeLine Dial %v", err)
		return nil, err
	}
	defer hc.Close()
	conn := redis.NewConn(hc, 0, 0) // TODO 0 TO's for now

	list, err := redis.Strings(conn.Do("LRANGE", "TL:"+userID, 0, -1))
	if err != nil && err != redis.ErrNil {
		cx.Errorf("GetTimeLine %v", err)
	}
	idx := 0
	find := func(vs []int, x int) (int, bool) {
		for i, v := range vs {
			if v == x {
				return i, true
			}
		}
		return 0, false
	}

	if lastid != "0" {
		for i := 0; i < len(list) && list[i] != lastid; i++ {

		}
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

		v, err := redis.Strings(conn.Do("HMGET", "IM:"+photoID, "date", userID, "flag"))
		if err != nil && err != redis.ErrNil {
			cx.Errorf("GetTimeLine HMGET %v", err)
		}
		if v[2] != "" {
			flags, err := strconv.Atoi(v[2])
			if err == nil && flags > 1 {
				continue // skip flag'd images
			}
		}
		likes, err := redis.Int(conn.Do("HLEN", "IM:"+photoID))
		if err != nil && err != redis.ErrNil {
			cx.Errorf("GetTimeLine HLEN %v", err)
		}
		s := strings.Split(photoID, ".")
		dn, err := redis.String(conn.Do("HGET", "HT:"+s[0], "dn"))
		if err != nil && err != redis.ErrNil {
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
	hc, err := socket.Dial(cx, "tcp", server)
	if err != nil {
		cx.Errorf("addUser Dial %v", err)
	}
	defer hc.Close()
	conn := redis.NewConn(hc, 0, 0) // TODO 0 TO's for now

	// See if we have done this already, block others.
	_, err = conn.Do("HSET", "HT:"+userID, "dn", displayName)
	if err != nil {
		cx.Errorf("addUser Exists %v", err)
	}
}

// like the user on redis
func like(cx appengine.Context, userID, photoID string) {
	hc, err := socket.Dial(cx, "tcp", server)
	if err != nil {
		cx.Errorf("like Dial %v", err)
		return
	}
	defer hc.Close()
	conn := redis.NewConn(hc, 0, 0) // TODO 0 TO's for now
	_, err = redis.Int(conn.Do("HSET", "IM:"+photoID, userID, "1"))
	if err != nil && err != redis.ErrNil {
		cx.Errorf("like %v", err)
	}
}

// unlike
func unlike(cx appengine.Context, userID, photoID string) {
	hc, err := socket.Dial(cx, "tcp", server)
	if err != nil {
		cx.Errorf("unlike Dial %v", err)
		return
	}
	defer hc.Close()
	conn := redis.NewConn(hc, 0, 0) // TODO 0 TO's for now
	_, err = redis.Int(conn.Do("HDEL", "IM:"+photoID, userID))
	if err != nil && err != redis.ErrNil {
		cx.Errorf("unlike %v", err)
	}
}

// flag will tell us that things may not be quite right with this image.
func flag(cx appengine.Context, userID, photoID string) {
	hc, err := socket.Dial(cx, "tcp", server)
	if err != nil {
		cx.Errorf("unlike Dial %v", err)
		return
	}
	defer hc.Close()
	conn := redis.NewConn(hc, 0, 0) // TODO 0 TO's for now
	_, err = redis.Int(conn.Do("HINCRBY", "IM:"+photoID, "flag", 1))
	if err != nil && err != redis.ErrNil {
		cx.Errorf("unlike %v", err)
	}
}
