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
//  URL='https://endpoints-dot-abelana-222.appspot.com/_ah/api/discovery/v1/apis/import/v1/rest'
//  curl -s $URL > import.rest.discovery
//  URL='https://endpoints-dot-abelana-222.appspot.com/_ah/api/discovery/v1/apis/timeline/v1/rest'
//  curl -s $URL > timeline.rest.discovery
//
// $ endpointscfg.py gen_client_lib java -bs gradle -o . management.rest.discovery
//   endpointscfg.py gen_client_lib java -bs gradle -o . timeline.rest.discovery
// iOS Apps
// Note the rpc suffix in the URL:
// $ URL='https://my-app-id.appspot.com/_ah/api/discovery/v1/apis/greeting/v1/rpc'
// $ curl -s $URL > greetings.rpc.discovery

type Comment struct {
	UserID string
	Text   string
}

type TLEntry struct {
	Date       time.Time
	UserID     string
	UsersPhoto string
	PhotoID    string
	Comments   []Comment
	likes      int
}

type Friend struct {
	UserID    string
	UserPhoto string
	Email     string
	Name      string
	ShareTo   bool
	ShareFrom bool
}

func init() {
	initTimeline()
	initImport()
	initFriend()
	initPhoto()
	initManagement()

	endpoints.HandleHttp()
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Timeline Service
///////////////////////////////////////////////////////////////////////////////////////////////////

// TimelineService will provide for GetTimeLine, GetMyTimeLine, and GetFriendTL

type TimelineService struct {
}

type TLReq struct {
	ATok string
}

type TLFReq struct {
	ATok   string
	friend string
}

type TLResp struct {
	resp []TLEntry
	Done string
}

func initTimeline() {
	timelineService := &TimelineService{}
	api, err := endpoints.RegisterService(timelineService, "Timeline", "v1", "Timeline API", true)
	if err != nil {
		panic(err.Error())
	}

	info := api.MethodByName("GetTimeLine").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"gettimeline", "GET", "timeline", "List of my main timeline."

	info = api.MethodByName("GetMyProfile").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"getmyprofile", "GET", "timeline", "List of my profile. (uploads)"

	info = api.MethodByName("GetFriendsProfile").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"getfriendsprofile", "GET", "timeline", "List of my friends profile. (uploads)"
}

func (ts *TimelineService) GetTimeLine(r *http.Request, req *TLReq, resp *TLResp) error {
	// Stub here
	return nil
}

func (ts *TimelineService) GetMyProfile(r *http.Request, req *TLReq, resp *TLResp) error {

	return ts.GetTimeLine(r, req, resp)
}

func (ts *TimelineService) GetFriendsProfile(r *http.Request, req *TLFReq, resp *TLResp) error {
	treq := new(TLReq)
	treq.ATok = req.ATok
	return ts.GetTimeLine(r, treq, resp)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Import Service
///////////////////////////////////////////////////////////////////////////////////////////////////
type ImportService struct {
}

type ImportReq struct {
	ATok  string
	Xcred string
}

func initImport() {
	importService := &ImportService{}
	api, err := endpoints.RegisterService(importService, "Import", "v1", "Import API", true)
	if err != nil {
		panic(err.Error())
	}
	info := api.MethodByName("Import").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"import", "PUT", "import", "Import friends from services"

}

func (gs *ImportService) Import(r *http.Request, req *ImportReq, resp *StatusResp) error {
	resp.Status = "Done"
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Friend Service
///////////////////////////////////////////////////////////////////////////////////////////////////
type FriendService struct {
}

type FLReq struct {
	ATok string
}

type FLResp struct {
	Friends []Friend
}

type FReq struct {
	ATok     string
	FriendID string
}

type FrReq struct {
	ATok     string
	FriendID string
	shareTo  bool
}

func initFriend() {
	friendService := &FriendService{}
	api, err := endpoints.RegisterService(friendService, "Friend", "v1", "Friend API", true)
	if err != nil {
		panic(err.Error())
	}
	info := api.MethodByName("GetFriendsList").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"getfriendslist", "GET", "friend", "Get Friends list for user"

}

func (gs *FriendService) GetFriendsList(r *http.Request, req *FLReq, resp *FLResp) error {

	return nil
}

func (gs *FriendService) GetFriend(r *http.Request, req *FReq, resp *Friend) error {

	return nil
}

func (gs *FriendService) SetFriend(r *http.Request, req *FrReq, resp *StatusResp) error {

	resp.Status = "Ok"
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Photo Service
///////////////////////////////////////////////////////////////////////////////////////////////////
type PhotoService struct {
}

type PhotoReq struct {
	ATok    string
	PhotoID string
}

func initPhoto() {
	//	photoService := &PhotoService{}
	//	api, err := endpoints.RegisterService(photoService, "Photos", "v1", "Photo API", true)
	//	if err != nil {
	//		panic(err.Error())
	//	}

}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Management Service
///////////////////////////////////////////////////////////////////////////////////////////////////
type ManagementService struct {
}

type LoginReq struct {
	GitTok string
}

type AToken struct {
	ATok string
}

type StatusResp struct {
	Status string
}

type GCMReq struct {
	ATok  string
	RegID string
}

type GCMResp struct {
	Done string
}

func initManagement() {
	managementService := &ManagementService{}
	api, err := endpoints.RegisterService(managementService, "Management", "v1", "Management API",
		true)
	if err != nil {
		panic(err.Error())
	}

	info := api.MethodByName("Login").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"login", "GET", "management", "Login and get a AccessToken token."

	info = api.MethodByName("Refresh").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"refresh", "GET", "management", "Refresh an AccessToken."

	info = api.MethodByName("Wipeout").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"wipeout", "DELETE", "management", "Wipeout users data."

	info = api.MethodByName("Register").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"register", "GET", "management", "Register device for GCM."

	info = api.MethodByName("Unregister").Info()
	info.Name, info.HttpMethod, info.Path, info.Desc =
		"unregister", "GET", "management", "Unregister device for GCM."

}

// Login
func (gs *ManagementService) Login(r *http.Request, req *LoginReq, resp *AToken) error {
	resp.ATok = "000001 ATOKEN - format will change significantly"
	return nil
}

// Refresh
func (gs *ManagementService) Refresh(r *http.Request, req *AToken, resp *AToken) error {
	resp.ATok = "000002 ATOKEN - format will change significantly"
	return nil
}

// Wipeout
func (gs *ManagementService) Wipeout(r *http.Request, req *AToken, resp *StatusResp) error {
	resp.Status = "Ok"
	return nil
}

// Register
func (gs *ManagementService) Register(r *http.Request, req *GCMReq, resp *GCMResp) error {
	resp.Done = "Ok"
	return nil
}

// Unregister
func (gs *ManagementService) Unregister(r *http.Request, req *GCMReq, resp *GCMResp) error {
	resp.Done = "Ok"
	return nil
}
