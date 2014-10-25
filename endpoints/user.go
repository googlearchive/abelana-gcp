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
	"io"
	"log"
	"net/http"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"

	"code.google.com/p/goauth2/oauth"

	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

// FindUser Lookup the user
func FindUser(cx appengine.Context, userID string) (*User, error) {
	cx.Infof("FindUser: %v", userID)
	user := &User{}
	err := datastore.Get(cx, datastore.NewKey(cx, "user", user.UserID, 0, nil), &user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateUser will create the initial datastore entry for the user
func CreateUser(cx appengine.Context, user *User) error {
	log.Printf("CreateUser: %v", user)
	_, err := datastore.Put(cx, datastore.NewKey(cx, "user", user.UserID, 0, nil), user)
	if err != nil {
		cx.Errorf(" CreateUser %v %v", err, user.UserID)
		return err
	}
	return nil
}

// CopyUserPhoto will copy the photo from
func CopyUserPhoto(cx appengine.Context, url string, userID string) error {
	// We want a larger photo
	url = strings.Replace(url, "sz=50", "sz=2048", 1)

	client := urlfetch.Client(cx)
	resp, err := client.Get(url)
	defer resp.Body.Close()
	if err != nil {
		cx.Errorf(" copyUserPhoto: %v %v %v", userID, url, err)
		return err
	}

	tok, _, err := appengine.AccessToken(cx, "https://www.googleapis.com/auth/devstorage.read_write")
	if err != nil {
		cx.Errorf(" AccessToken %v", err)
		return err
	}

	transport := &oauth.Transport{
		Token:     &oauth.Token{AccessToken: tok},
		Transport: &urlfetch.Transport{Context: cx},
	}
	clnt := &http.Client{Transport: transport}

	ctx := cloud.NewContext(projectID, clnt)
	w := storage.NewWriter(ctx, bucket, userID+".jpg", &storage.Object{ContentType: "image/jpg"})
	defer w.Close()

	_, err = io.Copy(w, resp.Body)

	if err := w.Close(); err != nil {
		cx.Errorf(" cup closing %v", err)
	}
	if _, err := w.Object(); err != nil {
		cx.Errorf("  .Object %v", err)
	}
	return err
}
