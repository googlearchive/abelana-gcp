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
func addPhoto(cx appengine.Context, superID string) {

	s := strings.Split(superID, ".")
	if len(s) != 3 {
		cx.Errorf("AddPhoto -- bad superID %v", superID)
		return
	}
	u, err := findUser(cx, s[0])
	if err != nil {
		cx.Errorf("AddPhoto unable to find user %v %v", superID, err)
		return
	}

	p := &Photo{s[1], time.Now().UTC().Unix()}
	hc, err := socket.Dial(cx, "tcp", server)
	if err != nil {
		cx.Errorf("AddPhoto Dial %v", err)
		return
	}
	defer hc.Close()
	conn := redis.NewConn(hc, 0, 0) // TODO 0 TO's for now

	photoID := s[0] + "." + s[1]
	set, err := redis.Int(conn.Do("HSETNX", "IM:"+photoID, "date", p.Date)) // Set Date
	if (err != nil && err != redis.ErrNil) || set == 0 {
		cx.Errorf("AddPhoto D duplicate %v %v", err, set)
		return // block duplicate requests
	}

	// Consider that at some point these should be done in batches of 100 or so. TODO

	// Add to each follower's list
	for _, personID := range u.Persons {
		conn.Send("LPUSH", "TL:"+personID, photoID)
	}
	conn.Flush()

	// Check the result and adjust list if nescessary.
	for _, personID := range u.Persons {
		v, err := redis.Int(conn.Receive())
		if err != nil && err != redis.ErrNil {
			cx.Errorf("AddPhoto for %v %v", "TL:"+personID, err)
		} else {
			if v > 2000 {
				_, err := conn.Do("RPOP", "TL:"+personID)
				if err != nil {
					cx.Errorf("AddPhoto RPOP %v", err)
				}
			}
		}
	}
	k1 := datastore.NewKey(cx, "User", s[0], 0, nil)
	k2 := datastore.NewKey(cx, "Photo", s[1], 0, k1)
	_, err = datastore.Put(cx, k2, p)
	if err != nil {
		cx.Errorf("AddPhoto datastore %v", err)
	}
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
