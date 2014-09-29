package launchEndpoints

import (
//    "fmt"
    "net/http"
    "time"
//    "appengine"
    "appengine/datastore"

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

// TimelineService will provide for GetTimeLine, GetMyTimeLine, and GetFriendTL
type TimelineService struct {
}

// Greeting is a datastore entity that represents a single greeting.
// It also serves as (a part of) a response of GreetingService.
type Greeting struct {
  Key     *datastore.Key `json:"id" datastore:"-"`
  Author  string         `json:"author"`
  Content string         `json:"content" datastore:",noindex" endpoints:"req"`
  Date    time.Time      `json:"date"`
}

// GreetingsList is a response type of GreetingService.List method
type GreetingsList struct {
  Items []*Greeting `json:"items"`
}

// Request type for GreetingService.List
type GreetingsListReq struct {
  Limit int `json:"limit" endpoints:"d=10"`
}


type ImportService struct {
}

type FriendService struct {
}

type PhotoService struct {
}

type ManagementService struct {
}

func init() {
  initTimeline()
  initImport()
  initFriend()
  initPhoto()
  initManagement()
  
  endpoints.HandleHttp()
}

// Timeline Service

func initTimeline() {
  timelineService := &TimelineService{}
  api, err := endpoints.RegisterService(timelineService, "Timeline", "v1", "Timeline API", true)
  if err != nil {
    panic(err.Error())
  }

  info := api.MethodByName("List").Info()
  info.Name, info.HttpMethod, info.Path, info.Desc =
    "greets.list", "GET", "greetings", "List most recent greetings."
}

//func (ts *TimelineService) GetTimeLine(r *http.Request, req *TLReq, resp *TLResp) error {
//
//    return nil
//}

// List responds with a list of all greetings ordered by Date field.
// Most recent greets come first.
func (gs *TimelineService) List(r *http.Request, req *GreetingsListReq, resp *GreetingsList) error {

//  if req.Limit <= 0 {
//    req.Limit = 10
//  }
//
//  c := endpoints.NewContext(r)

  return nil
}

// Import Service

func initImport() {
//  importService := &ImportService{}
//  api, err := endpoints.RegisterService(importService, "Import", "v1", "Import API", true)
//  if err != nil {
//    panic(err.Error())
//  }

}


// Friend Service

func initFriend() {
//  friendService := &FriendService{}
//  api, err := endpoints.RegisterService(importService, "Friend", "v1", "Friend API", true)
//  if err != nil {
//    panic(err.Error())
//  }

}

// Photo Service

func initPhoto() {

}

// Management Service

func initManagement() {

}

