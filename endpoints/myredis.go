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

	"github.com/GoogleCloudPlatform/abelana-gcp/third_party/redisx"

	"appengine"
	"appengine/datastore"
)

var (
	pool = redisx.Pool{
		MaxIdle:     3,
		IdleTimeout: 115 * time.Second,
		Dial: func(cx appengine.Context) (redisx.Conn, error) {
			c, err := redisx.Dial(cx, "tcp", abelanaConfig().Redis)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", abelanaConfig().RedisPW); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redisx.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
)


func newPool(server, password string) *redisx.Pool {
	return &redisx.Pool{
		MaxIdle:     3,
		IdleTimeout: 115 * time.Second,
		Dial: func(cx appengine.Context) (redisx.Conn, error) {
			c, err := redisx.Dial(cx, "tcp", abelanaConfig().Redis)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", abelanaConfig().RedisPW); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redisx.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// iNowFollow is Called when the user wants to follow someone
func iNowFollow(cx appengine.Context, userID, followerID string) {
	// TODO: implement
}

// addPhoto is called to add a photo. This is allways called from a Delay
func addPhoto(cx appengine.Context, photoID string) error {
	s := strings.Split(photoID, ".")

	// s[0] = userid, s[1] = random photo
	userID := s[0]
	u, err := findUser(cx, userID)
	if err != nil {
		return fmt.Errorf("addPhoto: unable to find user %v %v", userID, err)
	}

	p := &Photo{photoID, time.Now().UTC().Unix()}

	conn := pool.Get(cx)
	defer conn.Close()

	set, err := redisx.Int(conn.Do("HSETNX", "IM:"+photoID, "date", p.Date)) // Set Date
	if (err != nil && err != redisx.ErrNil) || set == 0 {
		return fmt.Errorf("addPhoto: duplicate %v %v", err, set)
	}

	// TODO: Consider if these should be done in batches of 100 or so.

	// Add to each follower's list
	for _, f := range u.FollowsMe {
		conn.Send("LPUSH", "TL:"+f, photoID)
	}
	conn.Flush()

	// Check the result and adjust list if nescessary.
	for _, f := range u.FollowsMe {
		v, err := redisx.Int(conn.Receive())
		if err != nil && err != redisx.ErrNil {
			cx.Errorf("addPhoto: get TL:%v %v", f, err)
			continue
		}
		if v > 2000 {
			if _, err := conn.Do("RPOP", "TL:"+f); err != nil {
				cx.Errorf("addPhoto: RPOP TL:%v %v", f, err)
			}
		}
	}
	k := datastore.NewKey(cx, "Photo", photoID, 0,
		datastore.NewKey(cx, "User", userID, 0, nil),
	)
	if _, err := datastore.Put(cx, k, p); err != nil {
		return fmt.Errorf("addPhoto: put photo in datastore %v", err)
	}
	return nil
}

// getTimeline returns the user's Timeline, you could insert additional things here as well.
func getTimeline(cx appengine.Context, userID, lastid string) ([]TLEntry, error) {
	conn := pool.Get(cx)
	defer conn.Close()

	list, err := redisx.Strings(conn.Do("LRANGE", "TL:"+userID, 0, -1))
	if err != nil && err != redisx.ErrNil {
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
	var timeline []TLEntry
	for i := 0; i < abelanaConfig().TimelineBatchSize && i+ix < len(list); i++ {
		photoID := list[ix+i]

		v, err := redisx.Strings(conn.Do("HMGET", "IM:"+photoID, "date", userID, "flag"))
		if err != nil && err != redisx.ErrNil {
			cx.Errorf("GetTimeLine HMGET %v", err)
		}
		if len(v) > 2 && v[2] != "" {
			flags, err := strconv.Atoi(v[2])
			if err == nil && flags > 1 {
				continue // skip flag'd images
			}
		}
		likes, err := redisx.Int(conn.Do("HLEN", "IM:"+photoID))
		if err != nil && err != redisx.ErrNil {
			cx.Errorf("GetTimeLine HLEN %v", err)
		}
		s := strings.Split(photoID, ".")
		dn, err := redisx.String(conn.Do("HGET", "HT:"+s[0], "dn"))
		if err != nil && err != redisx.ErrNil {
			cx.Errorf("GetTimeLine HLEN %v", err)
		}
		dt, err := strconv.ParseInt(v[0], 10, 64)
		te := TLEntry{dt, s[0], dn, photoID, likes - 1, v[1] == "1"}
		timeline = append(timeline, te)
	}
	return timeline, nil
}

// addUser adds the user to redis
func addUser(cx appengine.Context, id, name string) error {
	conn := pool.Get(cx)
	defer conn.Close()

	// See if we have done this already, block others.
	_, err := conn.Do("HSET", "HT:"+id, "dn", name)
	return err
}

// getPersons finds all the display names
func getPersons(cx appengine.Context, ids []string) ([]Person, error) {
	conn := pool.Get(cx)
	defer conn.Close()
	for _, id := range ids {
		conn.Send("HGET", "HT:"+id, "dn")
	}
	conn.Flush()

	var pl []Person
	for _, id := range ids {
		n, err := redisx.String(conn.Receive())
		if err != nil {
			return nil, fmt.Errorf("getPersons: receive name: %v", err)
		}
		pl = append(pl, Person{Kind: "abelana#follower", PersonID: id, Name: n})
	}
	return pl, nil
}

// like the user on redis
func like(cx appengine.Context, userID, photoID string) error {
	conn := pool.Get(cx)
	defer conn.Close()

	_, err := redisx.Int(conn.Do("HSET", "IM:"+photoID, userID, "1"))
	if err != nil && err != redisx.ErrNil {
		return fmt.Errorf("like %v", err)
	}
	return nil
}

// unlike
func unlike(cx appengine.Context, userID, photoID string) error {
	conn := pool.Get(cx)
	defer conn.Close()

	_, err := redisx.Int(conn.Do("HDEL", "IM:"+photoID, userID))
	if err != nil && err != redisx.ErrNil {
		return fmt.Errorf("unlike %v", err)
	}
	return nil
}

// flag will tell us that things may not be quite right with this image.
func flag(cx appengine.Context, userID, photoID string) error {
	conn := pool.Get(cx)
	defer conn.Close()

	_, err := redisx.Int(conn.Do("HINCRBY", "IM:"+photoID, "flag", 1))
	if err != nil && err != redisx.ErrNil {
		return fmt.Errorf("unlike %v", err)
	}
	return nil
}
