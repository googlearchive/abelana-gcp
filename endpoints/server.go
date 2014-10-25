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
	"strings"
	"time"

	"appengine"
	//    "appengine/datastore"
	"appengine/delay"
	"appengine/urlfetch"

	"github.com/go-martini/martini"
	// "github.com/garyburd/redigo/redis"
	auth "code.google.com/p/google-api-go-client/oauth2/v2"
	//	"github.com/golang/oauth2/google"
)

////////////////////////////////////////////////////////////////////
const EnableBackdoor = true // FIXME(lesv) TEMPORARY BACKDOOR ACCESS
////////////////////////////////////////////////////////////////////

const (
	authEmail     = "abelana-222@appspot.gserviceaccount.com"
	projectID     = "abelana-222"
	bucket        = "abelana-in"
	uploadRetries = 5
)

var delayFunc = delay.Func("test003", func(cx appengine.Context, x string) {
	cx.Infof("delay happened " + x)
})

// User is the root structure for everything.  For RockStars, it will probably get too large to
// memcache, so we'll skip that for now.
type User struct {
	UserID      string
	DisplayName string
	Email       string
	Friends     []string
}

// Comment holds all comments
type Comment struct {
	FriendID string `json:"friendid"`
	Text     string `json:"text"`
}

// Comments returned from GetComments()
type Comments struct {
	Kind    string    `json:"kind"`
	Entries []Comment `json:"entries"`
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
	kind     string `json:"kind"`
	FriendID string `json:"friendid"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// FriendList holds a list of our friends
type FriendList struct {
	kind    string   `json:"kind"`
	Friends []string `json:"friendid"`
}

// ATOKJson is the json message for an Access Token (TEMPORARY - Until GitKit supports this)
type ATOKJson struct {
	Kind string `json:"kind"`
	Atok string `json:"atok"`
}

// Status is what we return if we have nothing to return
type Status struct {
	Kind   string `json:"kind"`
	Status string `json:"status"`
}

// AppEngine middleware inserts a context where it's needed.
func AppEngine(c martini.Context, r *http.Request) {
	c.MapTo(appengine.NewContext(r), (*appengine.Context)(nil))
}

func init() {
	m := martini.Classic()
	m.Use(AppEngine)

	m.Get("/user/:gittok/login/:displayName/:photoUrl", Login)                  // => ATOKJson
	m.Get("/user/:atok/refresh", Aauth, Refresh)                                // => ATOKJson
	m.Get("/user/:atok/useful", Aauth, GetSecretKey)                            // => Status
	m.Delete("/user/:atok", Aauth, Wipeout)                                     // => Status
	m.Post("/user/:atok/facebook/:fbkey", Aauth, Import)                        // => Status
	m.Post("/user/:atok/plus/:plkey", Aauth, Import)                            // => Status
	m.Post("/user/:atok/yahoo/:ykey", Aauth, Import)                            // => Status
	m.Get("/user/:atok/friend", Aauth, GetFriendsList)                          // => FriendList
	m.Put("/user/:atok/friend/:friendid", Aauth, AddFriend)                     // => Status
	m.Get("/user/:atok/friend/:friendid", Aauth, GetFriend)                     // => Friend
	m.Put("/user/atok/device/:regid", Aauth, Register)                          // => Status
	m.Delete("/user/:atok/device/:regid", Aauth, Unregister)                    // => Status
	m.Get("/user/:atok/timeline/:lastid", Aauth, GetTimeLine)                   // => Timeline
	m.Get("/user/:atok/profile/:lastid", Aauth, GetMyProfile)                   // => Timeline
	m.Get("/user/:atok/friend/:friendid/profile/:lastid", Aauth, FriendProfile) // => Timeline
	m.Post("/photo/:atok/:photoid/comment/:text", Aauth, SetPhotoComments)      // => Status
	m.Put("/photo/:atok/:photoid/like", Aauth, Like)                            // => Status
	m.Delete("/photo/:atok/:photoid/like", Aauth, Unlike)                       // => Status
	m.Get("/photo/:atok/:photoid/flag", Aauth, Flag)                            // => Status
	m.Get("/photo/:atok/:photoid/comments", Aauth, GetComments)                 // => Comments
	m.Post("/photopush/:superid", PostPhoto)                                    // "ok"

	if EnableBackdoor {
		m.Get("/les", Test)
		m.Get("/user/:gittok/login", Login)
	}

	tokenInit()

	http.Handle("/", m)
}

// Test does magic of the moment
func Test(cx appengine.Context) string {
	cx.Infof("Test...")
	delayFunc.Call(cx, "hello world")
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

func replyOk(w http.ResponseWriter) {
	st := &Status{"abelana#status", "ok"}
	replyJSON(w, st)
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
func GetMyProfile(p martini.Params, w http.ResponseWriter, req *http.Request) {
	t := time.Now().Unix()
	timeline := []TLEntry{
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

// FriendProfile - Get a specific friends entries only (TlfReq) : TlResp
func FriendProfile(p martini.Params, w http.ResponseWriter, req *http.Request) {
	t := time.Now().Unix()
	timeline := []TLEntry{
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
		TLEntry{t - 99993, "00001", "0006", 99}}
	tl := &Timeline{"abelana#timeline", timeline}
	replyJSON(w, tl)
}

// PostPhoto lets us know that we have a photo, we then tell both DataStore and Redis
func PostPhoto(cx appengine.Context, p martini.Params, w http.ResponseWriter, rq *http.Request) string {
	cx.Infof("PostPhoto %v", p["superid"])
	otok := rq.Header.Get("Authorization")
	ok, err := authorized(cx, otok)
	if !ok {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return ``
	}
	return `ok`
}

func authorized(cx appengine.Context, token string) (bool, error) {
	if fs := strings.Fields(token); len(fs) == 2 && fs[0] == "Bearer" {
		token = fs[1]
	} else {
		return false, nil
	}

	svc, err := auth.New(urlfetch.Client(cx))
	if err != nil {
		return false, err
	}
	tok, err := svc.Tokeninfo().Access_token(token).Do()
	if err != nil {
		return false, err
	}
	cx.Infof("  tok %v", tok)
	return tok.Email == authEmail, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Import
///////////////////////////////////////////////////////////////////////////////////////////////////

// Import for Facebook / G+ / ... (xcred) : StatusResp
func Import(cx appengine.Context, p martini.Params, w http.ResponseWriter) {
	replyOk(w)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Friend
///////////////////////////////////////////////////////////////////////////////////////////////////

// GetFriendsList - A list of our friends (AToken) : FlResp
func GetFriendsList(cx appengine.Context, p martini.Params, w http.ResponseWriter) {
	fl := &FriendList{"abelana#friendList", []string{"00001", "12730648828453578083"}}
	replyJSON(w, fl)
}

// GetFriend -- find out about someone (FReq) : Friend
func GetFriend(cx appengine.Context, p martini.Params, w http.ResponseWriter) {
	f := &Friend{"abelana#friend", "00001", "lesv@abelana-app.com", "Les Vogel"}
	replyJSON(w, f)
}

// AddFriend - will tell us about a new possible friend (FrReq) : Status
func AddFriend(p martini.Params, w http.ResponseWriter) {
	replyOk(w)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Photo
///////////////////////////////////////////////////////////////////////////////////////////////////

// SetPhotoComments allows the users voice to be heard (PhotoComment) : Status
func SetPhotoComments(p martini.Params, w http.ResponseWriter) {
	replyOk(w)
}

// Like let's the user tell of their joy (Photo) : Status
func Like(p martini.Params, w http.ResponseWriter) {
	replyOk(w)
}

// Unlike let's the user recind their +1 (Photo) : Status
func Unlike(p martini.Params, w http.ResponseWriter) {
	replyOk(w)
}

// Flag will bring this to the administrators attention.
func Flag(p martini.Params, w http.ResponseWriter) {
	replyOk(w)
}

// GetComments will get the comments given a photoid
func GetComments(p martini.Params, w http.ResponseWriter) {
	cl := &Comments{"abelana#comments", []Comment{Comment{"00001", "Looks Great!"}}}
	replyJSON(w, cl)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Management
///////////////////////////////////////////////////////////////////////////////////////////////////

// Wipeout will erase all data you are working on. (Atok) : Status
func Wipeout(p martini.Params, w http.ResponseWriter) {
	replyOk(w)
}

// Register will start GCM messages to your device (GCMReq) : Status
func Register(p martini.Params, w http.ResponseWriter) {
	replyOk(w)
}

// Unregister will stop GCM messages from going to your device (GCMReq) : Status
func Unregister(p martini.Params, w http.ResponseWriter) {
	replyOk(w)
}
