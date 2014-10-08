package abelanaEndpoints

import (
	//    "fmt"
	"net/http"
	"time"
	//    "appengine"
	//    "appengine/datastore"

	"github.com/crhym3/go-endpoints/endpoints"
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
	Date       time.Time
	UserID     string
	UsersPhoto string
	PhotoID    string
	Comments   []Comment
	likes      int
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

// AToken is an Access Token - a series of base64 encoded strings concatenated together
type AToken struct {
	ATok string
}

// StatusResp - General status response
type StatusResp struct {
	Status string
}

// AbelanaService is used identify our endpoints
type AbelanaService struct{}

func init() {
	abelanaService := &AbelanaService{}
	api, err := endpoints.RegisterService(abelanaService, "Abelana", "v1", "Abelana API", true)
	if err != nil {
		panic(err.Error())
	}

	info := api.MethodByName("GetTimeLine").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"GetTimeline", "GET", "timeline", "List of my main timeline."

	info = api.MethodByName("GetMyProfile").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"GetMyProfile", "GET", "timeline/me", "List of my profile. (uploads)"

	info = api.MethodByName("GetFriendsProfile").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"GetFriendsProfile", "GET", "timeline/friend", "List of my friends profile. (uploads)"

	info = api.MethodByName("Import").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"Import", "PUT", "import", "Import friends from services"

	info = api.MethodByName("GetFriendsList").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"GetFriendsList", "GET", "friend/list", "Get Friends list for user"

	info = api.MethodByName("GetFriend").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"GetFriend", "GET", "friend", "Get a Friend"

	info = api.MethodByName("SetFriend").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"AddFriend", "GET", "friend/add", "Get Friends list for user"

	info = api.MethodByName("SetPhotoComments").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"SetComment", "PUT", "photo/setcomment", "Add comments to a photo"
	info = api.MethodByName("Like").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"Like", "PUT", "photo/like", "Mark a photo as liked"
	info = api.MethodByName("Unlike").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"Unlike", "PUT", "photo/unlike", "Remove me as having liked this photo"

	info = api.MethodByName("Login").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"Login", "GET", "management/login", "Login and get a AccessToken token."

	info = api.MethodByName("Refresh").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"Refresh", "GET", "management/refresh", "Refresh an AccessToken."

	info = api.MethodByName("Wipeout").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"Wipeout", "DELETE", "management/wipeout", "Wipeout users data."

	info = api.MethodByName("Register").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"RegisterGCM", "GET", "management/register", "Register device for GCM."

	info = api.MethodByName("Unregister").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"UnregisterGCM", "GET", "management/unregister", "Unregister device for GCM."

	endpoints.HandleHttp()
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
	resp []TLEntry
	Done string
}

// GetTimeLine - get the timeline for the user
func (as *AbelanaService) GetTimeLine(r *http.Request, req *AToken, resp *TlResp) error {
	// Stub here
	return nil
}

// GetMyProfile - Get my entries only
func (as *AbelanaService) GetMyProfile(r *http.Request, req *AToken, resp *TlResp) error {

	return as.GetTimeLine(r, req, resp)
}

// GetFriendsProfile - Get a specific friends entries only
func (as *AbelanaService) GetFriendsProfile(r *http.Request, req *TlfReq, resp *TlResp) error {
	treq := new(AToken)
	treq.ATok = req.ATok
	return as.GetTimeLine(r, treq, resp)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Import
///////////////////////////////////////////////////////////////////////////////////////////////////

// ImportReq Import Request
type ImportReq struct {
	ATok  string
	Xcred string
}

// Import for Facebook / G+ / ...
func (as *AbelanaService) Import(r *http.Request, req *ImportReq, resp *StatusResp) error {
	resp.Status = "Ok"
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Friend
///////////////////////////////////////////////////////////////////////////////////////////////////

// FlResp A list of friends
type FlResp struct {
	Friends []Friend
}

// FReq Request information about a friend
type FReq struct {
	ATok     string
	FriendID string
}

// FrReq Request sharing change
type FrReq struct {
	ATok     string
	FriendID string
	shareTo  bool
}

// GetFriendsList - A list of our friends
func (as *AbelanaService) GetFriendsList(r *http.Request, req *AToken, resp *FlResp) error {

	return nil
}

// GetFriend -- find out about someone
func (as *AbelanaService) GetFriend(r *http.Request, req *FReq, resp *Friend) error {

	return nil
}

// SetFriend - will tell us about a new possible friend
func (as *AbelanaService) SetFriend(r *http.Request, req *FrReq, resp *StatusResp) error {

	resp.Status = "Ok"
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Photo
///////////////////////////////////////////////////////////////////////////////////////////////////

// Photo is about a photo
type Photo struct {
	ATok    string
	PhotoID string
}

// PhotoComment -- What would you like to say
type PhotoComment struct {
	ATok    string
	PhotoID string
	Comment string
}

// SetPhotoComments allows the users voice to be heard
func (as *AbelanaService) SetPhotoComments(r *http.Request, req *PhotoComment, resp *StatusResp) error {

	resp.Status = "Ok"
	return nil
}

// Like let's the user tell of their joy
func (as *AbelanaService) Like(r *http.Request, req *Photo, resp *StatusResp) error {

	resp.Status = "Ok"
	return nil
}

// Unlike let's the user recind their +1
func (as *AbelanaService) Unlike(r *http.Request, req *Photo, resp *StatusResp) error {

	resp.Status = "Ok"
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Management
///////////////////////////////////////////////////////////////////////////////////////////////////

// LoginReq - Initial reqest for login
type LoginReq struct {
	GitTok string
}

// GCMReq - Google Cloud messaging request
type GCMReq struct {
	ATok  string
	RegID string
}

// GCMResp - Google Cloud messaging response
type GCMResp struct {
	Done string
}

// Login will validate the user and return an Access Token -- Null if invalid.
func (as *AbelanaService) Login(r *http.Request, req *LoginReq, resp *AToken) error {
	resp.ATok = "000001 ATOKEN - format will change significantly"
	return nil
}

// Refresh will refresh an Access Token
func (as *AbelanaService) Refresh(r *http.Request, req *AToken, resp *AToken) error {
	resp.ATok = "000002 ATOKEN - format will change significantly"
	return nil
}

// Wipeout will erase all data you are working on.
func (as *AbelanaService) Wipeout(r *http.Request, req *AToken, resp *StatusResp) error {
	resp.Status = "Ok"
	return nil
}

// Register will start GCM messages to your device
func (as *AbelanaService) Register(r *http.Request, req *GCMReq, resp *GCMResp) error {
	resp.Done = "Ok"
	return nil
}

// Unregister will stop GCM messages from going to your device
func (as *AbelanaService) Unregister(r *http.Request, req *GCMReq, resp *GCMResp) error {
	resp.Done = "Ok"
	return nil
}
