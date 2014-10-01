package launchEndpoints

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
// $ URL='https://my-app-id.appspot.com/_ah/api/discovery/v1/apis/greeting/v1/rest'
// $ curl -s $URL > greetings.rest.discovery
//
// Optionally check the discovery doc
// $ less greetings.rest.discovery
//
// $ GO_SDK/endpointscfg.py gen_client_lib java greetings.rest.discovery
//
// iOS Apps
// Note the rpc suffix in the URL:
// $ URL='https://my-app-id.appspot.com/_ah/api/discovery/v1/apis/greeting/v1/rpc'
// $ curl -s $URL > greetings.rpc.discovery
//
// Optionally check the discovery doc
// less greetings.rpc.discovery

type TLEntry struct {
    Date        time.Time
    UserID      string
    UsersPhoto  string
    PhotoID     string
    Comments    []string
    likes       []string
}

type Friend struct {

}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Management Service
///////////////////////////////////////////////////////////////////////////////////////////////////
type ManagementService struct {
}

type LoginReq struct {
  GitTok    string
}

type AToken struct {
  ATok      string
}

type StatusResp struct {
  Status string
}

type GCMReq struct {
    ATok    string
    RegID   string
}

type GCMResp struct {
    Done    string
}
  
///////////////////////////////////////////////////////////////////////////////////////////////////
// Timeline Service
///////////////////////////////////////////////////////////////////////////////////////////////////
// TimelineService will provide for GetTimeLine, GetMyTimeLine, and GetFriendTL
type TimelineService struct {
}

type TLReq struct {
    ATok    string
}

type TLFReq struct {
    ATok    string
    friend  string
}


type TLResp struct {
    resp    []TLEntry
    done    string
}


///////////////////////////////////////////////////////////////////////////////////////////////////
// Import Service
///////////////////////////////////////////////////////////////////////////////////////////////////
type ImportService struct {
}

type ImportReq struct {
    ATok    string
    Xcred   string
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Friend Service
///////////////////////////////////////////////////////////////////////////////////////////////////
type FriendService struct {
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Photo Service
///////////////////////////////////////////////////////////////////////////////////////////////////
type PhotoService struct {
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

    return nil
}

func (ts *TimelineService) GetMyProfile(r *http.Request, req *TLReq, resp *TLResp) error {

    return GetTimeLine(r, req, resp)
}

func (ts *TimelineService) GetFriendsProfile(r *http.Request, req *TLFReq, resp *TLResp) error {
  var rq TLReq
  rq.ATok = req.ATok
  return GetTimeLine(r, rq, resp)
}


///////////////////////////////////////////////////////////////////////////////////////////////////
// Import Service
///////////////////////////////////////////////////////////////////////////////////////////////////
func initImport() {
  importService := &ImportService{}
  api, err := endpoints.RegisterService(importService, "Import", "v1", "Import API", true)
  if err != nil {
    panic(err.Error())
  }
  info := api.MethodByName("Import").Info()
  info.Name, info.HttpMethod, info.Path, info.Desc =
    "import", "POST", "import", "Import friends from services"

}

func (gs *ImportService) Import(r *http.Request, req *ImportReq, resp *StatusResp) error {
    resp.Status = "Done"
    return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Friend Service
///////////////////////////////////////////////////////////////////////////////////////////////////

func initFriend() {
//  friendService := &FriendService{}
//  api, err := endpoints.RegisterService(importService, "Friend", "v1", "Friend API", true)
//  if err != nil {
//    panic(err.Error())
//  }

}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Photo Service
///////////////////////////////////////////////////////////////////////////////////////////////////

func initPhoto() {
//  friendService := &FriendService{}
//  api, err := endpoints.RegisterService(importService, "Friend", "v1", "Friend API", true)
//  if err != nil {
//    panic(err.Error())
//  }

}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Management Service
///////////////////////////////////////////////////////////////////////////////////////////////////

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
    resp.Status = "Done"
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
