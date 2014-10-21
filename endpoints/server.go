package abelanaEndpoints

import (
	//    "fmt"

	"encoding/json"
	"log"
	"net/http"
	"time"

	"appengine"
	//    "appengine/datastore"
	"github.com/go-martini/martini"
	// "github.com/garyburd/redigo/redis"
)

// The initial version of this will provide stubs for everything.  If there are changes, then
// be sure to get the Android app updated.  We are trying to put this together so that we can later
// refactor into several modules.

// Comment holds all comments
type Comment struct {
	UserID string
	Text   string
}

// TLEntry holds timeline entries
type TLEntry struct {
	Date    time.Time
	UserID  string
	PhotoID string
	Likes   int
}

type Timeline struct {
	Status  string
	Entries []TLEntry
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

func AppEngine(c martini.Context, r *http.Request) {
	c.MapTo(appengine.NewContext(r), (*appengine.Context)(nil))
}

func init() {
	m := martini.Classic()
	m.Use(AppEngine)

	log.Print("b4 things")

	m.Get("/login/:gittok", Login)

	m.Get("/user/:atok/refresh", AtokAuth, Refresh)
	m.Delete("/user/:atok", AtokAuth, Wipeout)
	m.Post("/user/:atok/facebook/:fbkey", AtokAuth, Import)
	m.Post("/user/:atok/plus/:plkey", AtokAuth, Import)
	m.Post("/user/:atok/yahoo/:ykey", AtokAuth, Import)
	m.Get("/user/:atok/friends/", AtokAuth, GetFriendsList)
	m.Get("/user/:atok/photo", AtokAuth, GetUserPhoto)
	m.Put("/user/:atok/friends/:friendid", AtokAuth, AddFriend)

	m.Get("/friend/:atok/:friendid", AtokAuth, GetFriend)

	m.Put("/device/:atok/:regid", AtokAuth, Register)
	m.Delete("/device/:atok/:regid", AtokAuth, Unregister)

	m.Get("/timeline/:atok/:lastid", AtokAuth, GetTimeLine)
	m.Get("/profile/:atok/:lastid", AtokAuth, GetMyProfile)
	m.Get("/profile/:atok/:userid/:lastid", AtokAuth, GetFriendsProfile)

	m.Post("/photo/:atok/:photoid/comment", AtokAuth, SetPhotoComments)
	m.Put("/photo/:atok/:photoid/like", AtokAuth, Like)
	m.Delete("/photo/:atok/:photoid/like", AtokAuth, Unlike)
	m.Get("/photo/:atok/:photoid/comments", AtokAuth, GetComments)

	tokenInit()

	http.Handle("/", m)
}

func Ok() string {
	return `{"Status": "Ok"}`
}

// replyJson Given an object, convert to JSON and reply with it
func replyJson(w http.ResponseWriter, v interface{}) {
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
	timeline := []TLEntry{
		TLEntry{time.Now(), "00001", "0001", 1},
		TLEntry{time.Now(), "00001", "0002", 99},
		TLEntry{time.Now(), "00001", "0003", 0},
		TLEntry{time.Now(), "00001", "0004", 3},
		TLEntry{time.Now(), "00001", "0005", 1},
		TLEntry{time.Now(), "00001", "0006", 99},
		TLEntry{time.Now(), "00001", "0007", 0},
		TLEntry{time.Now(), "00001", "0008", 3},
		TLEntry{time.Now(), "00001", "0009", 1},
		TLEntry{time.Now(), "00001", "0010", 99},
		TLEntry{time.Now(), "00001", "0011", 0},
		TLEntry{time.Now(), "00001", "0004", 3},
		TLEntry{time.Now(), "00001", "0001", 1},
		TLEntry{time.Now(), "00001", "0002", 99},
		TLEntry{time.Now(), "00001", "0003", 0},
		TLEntry{time.Now(), "00001", "0005", 3},
		TLEntry{time.Now(), "00001", "0007", 1},
		TLEntry{time.Now(), "00001", "0006", 99},
		TLEntry{time.Now(), "00001", "0011", 0},
		TLEntry{time.Now(), "00001", "0009", 3},
		TLEntry{time.Now(), "00001", "0002", 99},
		TLEntry{time.Now(), "00001", "0003", 0},
		TLEntry{time.Now(), "00001", "0004", 3},
		TLEntry{time.Now(), "00001", "0005", 1},
		TLEntry{time.Now(), "00001", "0006", 99},
		TLEntry{time.Now(), "00001", "0007", 0},
		TLEntry{time.Now(), "00001", "0008", 3},
		TLEntry{time.Now(), "00001", "0009", 1},
		TLEntry{time.Now(), "00001", "0010", 99},
		TLEntry{time.Now(), "00001", "0011", 0},
		TLEntry{time.Now(), "00001", "0004", 3},
		TLEntry{time.Now(), "00001", "0001", 1},
		TLEntry{time.Now(), "00001", "0002", 99},
		TLEntry{time.Now(), "00001", "0003", 0},
		TLEntry{time.Now(), "00001", "0005", 3},
		TLEntry{time.Now(), "00001", "0007", 1},
		TLEntry{time.Now(), "00001", "0006", 99},
		TLEntry{time.Now(), "00001", "0011", 0},
		TLEntry{time.Now(), "00001", "0009", 3},
		TLEntry{time.Now(), "00001", "0005", 21}}
	tl := &Timeline{"Ok", timeline}
	replyJson(w, tl)
}

// GetMyProfile - Get my entries only (token) : TlResp
func GetMyProfile(p martini.Params) string {
	return Ok()
}

// GetFriendsProfile - Get a specific friends entries only (TlfReq) : TlResp
func GetFriendsProfile(p martini.Params) string {
	return Ok()
}

func GetUserPhoto(p martini.Params) string {
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
