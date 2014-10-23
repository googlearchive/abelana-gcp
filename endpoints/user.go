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
	"errors"
	"log"
	"strings"

	"appengine/urlfetch"

	"appengine"
)

type User struct {
	UserId      string
	FirstName   string
	LastName    string
	DisplayName string
	Email       string
	Friends     []string
}

// FindUser Lookup the user
func FindUser(cx appengine.Context, userId string) (*User, error) {

	log.Printf("FindUser: %v", userId)
	return nil, errors.New("not found")
}

// CreateUser will create the initial datastore entry for the user
func CreateUser(cx appengine.Context, user *User) {
	log.Printf("CreateUser: %v", user)

}

// CopyUserPhoto will copy the photo from
func CopyUserPhoto(cx appengine.Context, url string, userId string) {
	// We want a larger photo
	url = strings.Replace(url, "sz=50", "sz=2048", 1)
	log.Printf("CopyUserPhoto: %v %v", userId, url)
	client := urlfetch.Client(cx)
	resp, err := client.Get(url)
	if err == nil {
		log.Printf("CopyUserPhto: (%v) %v %v", resp.StatusCode, err, resp.ContentLength)

		resp.Body.Close()
	}
}
