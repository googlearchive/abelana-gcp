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
	"net/http"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"

	"code.google.com/p/goauth2/oauth"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

// findUser Lookup the user (This can be called from a Transaction)
func findUser(cx appengine.Context, id string) (*User, error) {
	key := datastore.NewKey(cx, "User", id, 0, nil)
	cx.Infof("FindUser: %v", id)
	cx.Infof("  Key=%v", key)

	u := &User{}
	err := datastore.Get(cx, key, u)
	return u, err
}

// createUser will create the initial datastore entry for the user
func createUser(cx appengine.Context, user User) error {
	cx.Infof("CreateUser: %v", user)
	_, err := datastore.Put(cx, datastore.NewKey(cx, "User", user.UserID, 0, nil), &user)
	if err != nil {
		cx.Errorf(" CreateUser %v %v", err, user.UserID)
		return err
	}
	addUser(cx, user.UserID, user.DisplayName)
	return nil
}

// copyUserPhoto will copy the photo from, will likey be called from delayFunc
func copyUserPhoto(cx appengine.Context, url string, userID string) error {
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

	ctx := cloud.NewContext(abelanaConfig().ProjectID, clnt)
	w := storage.NewWriter(ctx, abelanaConfig().Bucket, userID+".jpg", &storage.Object{ContentType: "image/jpg"})
	defer w.Close()

	_, err = io.Copy(w, resp.Body)

	if err := w.Close(); err != nil {
		cx.Errorf(" cup closing %v", err)
	}
	if _, err := w.Object(); err != nil {
		cx.Errorf("  .Object %v", err)
	}
	cx.Infof("CopyUserPhoto ok %v %v", userID, url)
	return nil
}
