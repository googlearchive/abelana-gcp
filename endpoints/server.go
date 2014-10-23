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
	//    "fmt"

	"encoding/json"
	"net/http"
	"time"

	"appengine/user"

	"appengine"
	//    "appengine/datastore"
	"github.com/go-martini/martini"
	// "github.com/garyburd/redigo/redis"
)

// Comment holds all comments
type Comment struct {
	UserID string
	Text   string
}

// TLEntry holds timeline entries
type TLEntry struct {
	Created int64  `json:"created"`
	UserID  string `json:"userid"`
	PhotoID string `json:"photoid"`
	Likes   int    `json:"likes"`
}

// Timeline the data the client sees.
type Timeline struct {
	Kind    string    `json:"kind"`
	Entries []TLEntry `json:"entries"`
}

// Friend holds information about our friends
type Friend struct {
	UserID    string
	UserPhoto string
	Email     string
	Name      string
	ShareTo   bool
	ShareFrom bool
}

// AppEngine middleware inserts a context where it's needed.
func AppEngine(c martini.Context, r *http.Request) {
	c.MapTo(appengine.NewContext(r), (*appengine.Context)(nil))
}

func init() {
	m := martini.Classic()
	m.Use(AppEngine)

	m.Get("/user/:gittok/login/:displayName/:photoUrl", Login)
	m.Get("/user/:gittok/login", Login)

	m.Get("/user/:atok/refresh", AtokAuth, Refresh)
	m.Delete("/user/:atok", AtokAuth, Wipeout)
	m.Post("/user/:atok/facebook/:fbkey", AtokAuth, Import)
	m.Post("/user/:atok/plus/:plkey", AtokAuth, Import)
	m.Post("/user/:atok/yahoo/:ykey", AtokAuth, Import)
	m.Get("/user/:atok/photo", AtokAuth, GetUserPhoto)

	m.Get("/user/:atok/friend", AtokAuth, GetFriendsList)
	m.Put("/user/:atok/friend/:friendid", AtokAuth, AddFriend)
	m.Get("/user/:atok/friend/:friendid", AtokAuth, GetFriend)

	m.Put("/user/atok/device/:regid", AtokAuth, Register)
	m.Delete("/user/:atok/device/:regid", AtokAuth, Unregister)

	m.Get("/user/:atok/timeline/:lastid", AtokAuth, GetTimeLine)
	m.Get("/user/:atok/profile/:lastid", AtokAuth, GetMyProfile)
	m.Get("/user/:atok/friend/:friendid/profile/:lastid", AtokAuth, GetFriendsProfile)

	m.Post("/photo/:atok/:photoid/comment", AtokAuth, SetPhotoComments)
	m.Put("/photo/:atok/:photoid/like", AtokAuth, Like)
	m.Delete("/photo/:atok/:photoid/like", AtokAuth, Unlike)
	m.Get("/photo/:atok/:photoid/comments", AtokAuth, GetComments)

	m.Post("/photopush/:superid", PostPhoto)

	tokenInit()

	http.Handle("/", m)
}

// Ok simple reply for string versions
func Ok() string {
	return `ok`
}

// replyJSON Given an object, convert to JSON and reply with it
func replyJSON(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Add("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Timeline
///////////////////////////////////////////////////////////////////////////////////////////////////

// GetTimeLine - get the timeline for the user (token) : TlResp
func GetTimeLine(p martini.Params, w http.ResponseWriter, req *http.Request) {
	t := time.Now().Unix()
	timeline := []TLEntry{
		TLEntry{t - 200, "00001", "0001", 1},
		TLEntry{t - 1000, "00001", "0002", 99},
		TLEntry{t - 2500, "00001", "0003", 0},
		TLEntry{t - 6040, "00001", "0004", 3},
		TLEntry{t - 7500, "00001", "0005", 1},
		TLEntry{t - 9300, "00001", "0006", 99},
		TLEntry{t - 10200, "00001", "0007", 0},
		TLEntry{t - 47003, "00001", "0008", 3},
		TLEntry{t - 53002, "00001", "0009", 1},
		TLEntry{t - 54323, "00001", "0010", 99},
		TLEntry{t - 56112, "00001", "0011", 0},
		TLEntry{t - 58243, "00001", "0004", 3},
		TLEntry{t - 80201, "00001", "0001", 1},
		TLEntry{t - 80500, "00001", "0002", 99},
		TLEntry{t - 81200, "00001", "0003", 0},
		TLEntry{t - 89302, "00001", "0005", 3},
		TLEntry{t - 91200, "00001", "0007", 1},
		TLEntry{t - 92343, "00001", "0006", 99},
		TLEntry{t - 93233, "00001", "0011", 0},
		TLEntry{t - 94322, "00001", "0009", 3},
		TLEntry{t - 95323, "00001", "0002", 99},
		TLEntry{t - 96734, "00001", "0003", 0},
		TLEntry{t - 98033, "00001", "0004", 3},
		TLEntry{t - 99334, "00001", "0005", 1},
		TLEntry{t - 99993, "00001", "0006", 99},
		TLEntry{t - 102304, "00001", "0007", 0},
		TLEntry{t - 102750, "00001", "0008", 3},
		TLEntry{t - 104333, "00001", "0009", 1},
		TLEntry{t - 105323, "00001", "0010", 99},
		TLEntry{t - 107323, "00001", "0011", 0},
		TLEntry{t - 109323, "00001", "0004", 3},
		TLEntry{t - 110000, "00001", "0001", 1},
		TLEntry{t - 110133, "00001", "0002", 99},
		TLEntry{t - 113444, "00001", "0003", 0},
		TLEntry{t - 122433, "00001", "0005", 3},
		TLEntry{t - 125320, "00001", "0007", 1},
		TLEntry{t - 125325, "00001", "0006", 99},
		TLEntry{t - 127555, "00001", "0011", 0},
		TLEntry{t - 128333, "00001", "0009", 3},
		TLEntry{t - 173404, "00001", "0005", 21}}
	tl := &Timeline{"abelana#timeline", timeline}
	replyJSON(w, tl)
}

// GetMyProfile - Get my entries only (token) : TlResp
func GetMyProfile(p martini.Params) string {
	return Ok()
}

// GetFriendsProfile - Get a specific friends entries only (TlfReq) : TlResp
func GetFriendsProfile(p martini.Params) string {
	return Ok()
}

// GetUserPhoto Will return the user photo It should just be userID.webm
// (ie, we won't be needing this function)
func GetUserPhoto(p martini.Params) string {
	return Ok()
}

// PostPhoto lets us know that we have a photo, we then tell both DataStore and Redis
func PostPhoto(cx appengine.Context, p martini.Params, w http.ResponseWriter, rq *http.Request) string {
	cx.Infof("PostPhoto %v", p["superid"])
	u, err := user.CurrentOAuth(cx, "")
	cx.Infof("pp %v", u)
	if err != nil {
		cx.Infof("Oauth unauthorized %v", err)
		cx.Infof(" rq %v", rq.Header)
		http.Error(w, "OAuth Authorization header required", http.StatusUnauthorized)
		return ""
	}

	// if !u.Admin {
	// 	http.Error(w, "Admin login only", http.StatusUnauthorized)
	// 	return ""
	// }
	return Ok()
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Import
///////////////////////////////////////////////////////////////////////////////////////////////////

// Import for Facebook / G+ / ... (xcred) : StatusResp
func Import(p martini.Params) string {
	return Ok()
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Friend
///////////////////////////////////////////////////////////////////////////////////////////////////

// FlResp A list of friends
type FlResp struct {
	Friends []Friend
}

// FrReq Request sharing change
type FrReq struct {
	ATok     string
	FriendID string
	shareTo  bool
}

// GetFriendsList - A list of our friends (AToken) : FlResp
func GetFriendsList(p martini.Params) string {
	return Ok()
}

// GetFriend -- find out about someone (FReq) : Friend
func GetFriend(p martini.Params) string {
	return Ok()
}

// AddFriend - will tell us about a new possible friend (FrReq) : Status
func AddFriend(p martini.Params) string {
	return Ok()
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Photo
///////////////////////////////////////////////////////////////////////////////////////////////////

// SetPhotoComments allows the users voice to be heard (PhotoComment) : Status
func SetPhotoComments(p martini.Params) string {
	return Ok()
}

// Like let's the user tell of their joy (Photo) : Status
func Like(p martini.Params) string {
	return Ok()
}

// Unlike let's the user recind their +1 (Photo) : Status
func Unlike(p martini.Params) string {
	return Ok()
}

// GetComments will get the comments given a photoid
func GetComments(p martini.Params) string {
	return `{"Status": "Ok", "Comments":[]}`
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Management
///////////////////////////////////////////////////////////////////////////////////////////////////

// Refresh will refresh an Access Token (ATok)
func Refresh(p martini.Params) string {
	return `{"Status": "Ok", "Atok": "LES002"}`
}

// Wipeout will erase all data you are working on. (Atok) : Status
func Wipeout(p martini.Params) string {
	return Ok()
}

// Register will start GCM messages to your device (GCMReq) : Status
func Register(p martini.Params) string {
	return Ok()
}

// Unregister will stop GCM messages from going to your device (GCMReq) : Status
func Unregister(p martini.Params) string {
	return Ok()
}
