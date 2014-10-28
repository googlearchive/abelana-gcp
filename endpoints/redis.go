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

// AddTheUser to redis
func AddTheUser(cx appengine.Context, userID string) {

	hc, err := socket.Dial(cx, "tcp", server)
	if err != nil {
		cx.Errorf("AddTheUser Dial %v", err)
	}
	defer hc.Close()
	conn := redis.NewConn(hc, 0, 0) // TODO 0 TO's for now

	// See if we have done this already, block others.
	_, err = redis.Bool(conn.Do("SET", userID, "1", "NX"))
	if err != nil {
		cx.Errorf("AddTheUser Exists %v", err)
		return
	}
}

// addTheFolower is Called when we need to add a folower
func addTheFollower(cx appengine.Context, userID, followerID string) {

}

// AddPhoto is called to add a photo.
func AddPhoto(cx appengine.Context, superID string) {

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

	photo := s[0] + "." + s[1]
	ok, err := redis.String(conn.Do("SET", photo, "0", "NX")) // Set Likes to 0 if new
	if err != nil && err != redis.ErrNil {
		cx.Errorf("AddPhoto Exists %v", err)
		return
	}
	if ok != "OK" {
		cx.Errorf("AddPhoto duplicate %v", ok)
		return
	}

	ok, err = redis.String(conn.Do("SET", photo+"D", p.Date)) // Set Likes to 0 if new
	if err != nil && err != redis.ErrNil {
		cx.Errorf("AddPhoto D Exists %v", err)
		return
	}
	if ok != "OK" {
		cx.Errorf("AddPhoto D duplicate %v", ok)
		return
	}

	ok, err = redis.String(conn.Do("SET", photo+"N", u.DisplayName)) // Set Likes to 0 if new
	if err != nil && err != redis.ErrNil {
		cx.Errorf("AddPhoto N Exists %v", err)
		return
	}
	if ok != "OK" {
		cx.Errorf("AddPhoto N duplicate %v", ok)
		return
	}

	// TODO -- at some point these should be done in batches of 100 or so.
	for _, followerID := range u.Followers {
		conn.Send("LPUSH", followerID+"P", photo)
	}
	conn.Flush()
	for _, followerID := range u.Followers {
		v, err := redis.Int(conn.Receive())
		if err != nil && err != redis.ErrNil {
			cx.Errorf("AddPhoto for %v %v", followerID+"P", err)
		} else {
			if v > 2000 {
				_, err := conn.Do("RPOP", followerID)
				if err != nil {
					cx.Errorf("AddPhoto RPOP %v", err)
				}
			}
		}
	}
	k1 := datastore.NewKey(cx, "User", s[0], 0, nil)
	k2 := datastore.NewKey(cx, "Photo", s[1], 0, k1)
	_, err = datastore.Put(cx, k2, p)
}
