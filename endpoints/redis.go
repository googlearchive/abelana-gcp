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

	"appengine"
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

// AddTheFriend is Called when we need to add a friend
func AddTheFriend(cx appengine.Context, userID, friendID string) {

}

// AddPhoto is called to add a photo, at the moment, there isn't anything other than the name, so
// why bother adding to the datastore.  We'll store comments as UserID > PhotoID > CommentID
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
	hc, err := socket.Dial(cx, "tcp", server)
	if err != nil {
		cx.Errorf("AddPhoto Dial %v", err)
		return
	}
	defer hc.Close()
	conn := redis.NewConn(hc, 0, 0) // TODO 0 TO's for now

	photo := s[0] + "." + s[1]
	// See if we have done this already, block others.
	ok, err := redis.String(conn.Do("SET", photo, "1", "NX"))
	if err != nil && err != redis.ErrNil {
		cx.Errorf("AddPhoto Exists %v", err)
		return
	}
	if ok != "OK" {
		cx.Errorf("AddPhoto duplicate %v", ok)
		return
	}
	for _, friendID := range u.Friends {
		conn.Send("LPUSH", friendID+"P", photo)
	}
	conn.Flush()
	for _, friendID := range u.Friends {
		v, err := redis.Int(conn.Receive())
		if err != nil && err != redis.ErrNil {
			cx.Errorf("AddPhoto for %v %v", friendID+"P", err)
		} else {
			if v > 2000 {
				_, err := conn.Do("RPOP", friendID)
				if err != nil {
					cx.Errorf("AddPhoto RPOP %v", err)
				}
			}
		}
	}
}
