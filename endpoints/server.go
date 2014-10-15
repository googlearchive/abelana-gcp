package abelanaEndpoints

import (
	//    "fmt"

	"encoding/json"
	"log"
	"net/http"
	"time"
	// "appengine"
	//    "appengine/datastore"
	//	"github.com/cloud-abelana-go/appengine/endpoints/tokens"
	"github.com/go-martini/martini"
)

// The initial version of this will provide stubs for everything.  If there are changes, then
// be sure to get the Android app updated.  We are trying to put this together so that we can later
// refactor into several modules.

// How to generate Discovery Documentation & Client Liraries
//
// Android Apps
//  URL='https://endpoints-dot-abelana-222.appspot.com/_ah/api/discovery/v1/apis/management/v1/rest'
//  curl -s $URL > management.rest.discovery
//
// $ endpointscfg.py gen_client_lib java -bs gradle -o . management.rest.discovery
// iOS Apps
// Note the rpc suffix in the URL:
// $ URL='https://my-app-id.appspot.com/_ah/api/discovery/v1/apis/greeting/v1/rpc'
// $ curl -s $URL > greetings.rpc.discovery

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

// func CustomMartini() *martini.ClassicMartini {
// 	r := martini.NewRouter()
// 	m := martini.New()
// 	m.Use(martini.Logger())
// 	m.Use(martini.Recovery())
// 	m.Use(martini.Static("public"))
// 	m.MapTo(r, (*martini.Routes)(nil))
// 	m.Action(r.Handle)
// 	return &martini.ClassicMartini{m, r}
// }

// func AppEngine(c appengine.Context, r *http.Request) {
// 	c.Map(appengine.NewContext(r))
// }

func init() {
	m := martini.Classic()
	// m.MapTo(appengine.NewContext(r), (*appengine.Context)(nil))

	// m.Use(func(c martini.Context, req *http.Request) {
	// 	ctx := appengine.NewContext(req)
	// 	c.MapTo(ctx, (*appengine.Context)(nil))
	// })

	m.Get("/", func() string {
		return "Hello from Martini/server!"
	})
	m.Get("/hello/:name", func(p martini.Params) string {
		return "Hi there " + p["name"]
	})
	log.Print("b4 things")

	m.Get("/login/:gittok", Login)

	// m.Use(AuthMe) // Hopefully, only do this for atok's

	// The basics of REST
	// When dealing with a Collection URI like: http://example.com/resources/
	// GET: List the members of the collection, complete with their member URIs for further navigation.
	// For example, list all the cars for sale.
	// PUT: Meaning defined as "replace the entire collection with another collection".
	// POST: Create a new entry in the collection where the ID is assigned automatically by the
	// collection. The ID created is usually included as part of the data returned by this operation.
	// DELETE: Meaning defined as "delete the entire collection".
	//
	// When dealing with a Member URI like: http://example.com/resources/7HOU57Y
	// GET: Retrieve a representation of the addressed member of the collection expressed in an
	// appropriate MIME type.
	// PUT: Update the addressed member of the collection or create it with the specified ID.
	// POST: Treats the addressed member as a collection in its own right and creates a new subordinate
	// of it.
	// DELETE: Delete the addressed member of the collection.

	m.Get("/user/:atok/refresh", Refresh)
	m.Delete("/user/:atok", Wipeout)
	m.Post("/user/:atok/facebook/:fbkey", Import)
	m.Post("/user/:atok/plus/:plkey", Import)
	m.Post("/user/:atok/yahoo/:ykey", Import)
	m.Get("/user/:atok/friends/", GetFriendsList)
	m.Get("/user/:atok/photo", GetUserPhoto)
	m.Put("/user/:atok/friends/:friendid", AddFriend)

	m.Get("/friend/:atok/:friendid", GetFriend)

	m.Put("/device/:atok/:regid", Register)
	m.Delete("/device/:atok/:regid", Unregister)

	m.Get("/timeline/:atok/:lastid", GetTimeLine)
	m.Get("/profile/:atok/:lastid", GetMyProfile)
	m.Get("/profile/:atok/:userid/:lastid", GetFriendsProfile)

	m.Post("/photo/:atok/:photoid/comment", SetPhotoComments)
	m.Put("/photo/:atok/:photoid/like", Like)
	m.Delete("/photo/:atok/:photoid/like", Unlike)
	m.Get("/photo/:atok/:photoid/comments", GetComments)

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

// TlfReq Timeline Friend Request
type TlfReq struct {
	ATok   string
	friend string
}

// TlResp Response timeline
type TlResp struct {
	resp   []TLEntry
	Status string
}

// GetTimeLine - get the timeline for the user (token) : TlResp
func GetTimeLine(p martini.Params, w http.ResponseWriter, req *http.Request) {
	timeline := []TLEntry{
		TLEntry{time.Now(), "00001", "0001", 1},
		TLEntry{time.Now(), "00001", "0002", 99},
		TLEntry{time.Now(), "00001", "0003", 0},
		TLEntry{time.Now(), "00001", "0004", 3},
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

// Login will validate the user and return an Access Token -- Null if invalid. (LoginReq) : ATok
func Login(p martini.Params) string {
	return `{"Status": "Ok", "Atok": "LES001"}`
}

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
