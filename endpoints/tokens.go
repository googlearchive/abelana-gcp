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

// Package abelana is a set of utilities to validate our GitKit and Access Tokens.  For now, we are
// providing our own Access Tokens, later, we will use GitKit's tokens when they become available.
package abelana

import (
	"crypto/ecdsa"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"appengine"

	"github.com/go-martini/martini"
	"github.com/google/identity-toolkit-go-client/gitkit"
)

var (
	gclient *gitkit.Client
	signKey *ecdsa.PrivateKey
)

func init() {
	var config *gitkit.Config
	var err error
	// Provide configuration. gitkit.LoadConfig() can also be used to load
	// the configuration from a JSON file.
	if appengine.IsDevAppServer() {
		config, err = gitkit.LoadConfig("private/gitkit-server-config-dev.json")
	} else {
		config, err = gitkit.LoadConfig("private/gitkit-server-config.json")
	}
	if err != nil {
		log.Fatalf("Unable to initialize gitkit config %v", err)
	}
	gclient, err = gitkit.New(config)
	if err != nil {
		log.Fatalf("new gitkit.New ** %v", err)
	}
	key, err := ioutil.ReadFile("private/signing-key.pem")
	if err != nil {
		log.Fatalf("Unable to get signing Key %v", err)
	}
	b, _ := pem.Decode(key)
	signKey, err = x509.ParseECPrivateKey(b.Bytes)
	if err != nil {
		log.Fatalf("unable to parse signing Key %v", err)
	}
}

// Login - see if the token is valid
func Login(cx appengine.Context, p martini.Params, w http.ResponseWriter) {
	var token *gitkit.Token
	var err error
	var dName, photoURL string

	client, err := gitkit.NewWithContext(cx, gclient)
	if err != nil {
		cx.Errorf("Failed to create a gitkit.Client with a context: %s", err)
		http.Error(w, "Initialization failure", http.StatusInternalServerError)
		return
	}
	dn, err := decodeSegment(p["displayName"])
	if err != nil {
		dName = "Name Unavailable"
	} else {
		dName = string(dn)
	}
	pu, err := decodeSegment(p["photoUrl"])
	if err != nil {
		photoURL = ""
	} else {
		photoURL = string(pu)
	}
	if abelanaConfig().EnableBackdoor && p["gittok"] == "Les" {
		err = nil
		token = &gitkit.Token{"Magic", "**AUDIENCE**", time.Now().UTC(),
			time.Now().UTC().Add(1 * time.Hour), "00001", "lesv@abelana-app.com",
			true, "abelana-app.com", "LES001"}
		dName = "Les Vogel"
		photoURL = "https://lh4.googleusercontent.com/-Nt9PfYHmQeI/AAAAAAAAAAI/AAAAAAAAANI/2mbohwDXFKI/photo.jpg?sz=50"
	} else {
		token, err = client.ValidateToken(p["gittok"])
		if err != nil {
			cx.Errorf("git.ValidateToken: %v")
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
	}

	at := &AccToken{token.LocalID, time.Now().UTC().Unix(), time.Now().UTC().Add(120 * 24 * time.Hour).Unix()}

	parts := make([]string, 3)

	parts[0] = base64.URLEncoding.EncodeToString([]byte(`{"kid": "abelana"}`))
	ts, err := json.Marshal(at)
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}
	parts[1] = base64.URLEncoding.EncodeToString(ts)

	h := md5.New()
	io.WriteString(h, parts[0]+"."+parts[1])
	r, s, err := ecdsa.Sign(rand.Reader, signKey, h.Sum(nil))
	if err != nil {
		cx.Errorf("ecdsa.Sign %v", err)
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}
	sig := base64.URLEncoding.EncodeToString(r.Bytes()) + "." + base64.URLEncoding.EncodeToString(s.Bytes())
	parts[2] = base64.URLEncoding.EncodeToString([]byte(sig))

	replyJSON(w, &ATOKJson{"abelana#accessToken", strings.Join(parts, ".")})

	// Look us up in datastore and be happy.
	_, err = findUser(cx, at.UserID)
	if err != nil {
		// Not found, must create
		createUser(cx, User{UserID: at.UserID, DisplayName: dName, Email: token.Email})
		if photoURL != "" && photoURL != "null" {
			delayCopyUserPhoto.Call(cx, photoURL, at.UserID)
		}
	}
}

// Refresh will refresh an Access Token (ATok)
func Refresh(cx appengine.Context, p martini.Params, w http.ResponseWriter) {
	//	haveCerts(cx)
	parts := strings.Split(p["atok"], ".")
	ct, err := base64.URLEncoding.DecodeString(parts[1])
	at := &AccToken{}
	if err = json.Unmarshal(ct, &at); err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}
	at.Exp = time.Now().UTC().Add(120 * 24 * time.Hour).Unix()

	ts, err := json.Marshal(at)
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}
	parts[1] = base64.URLEncoding.EncodeToString(ts)

	h := md5.New()
	io.WriteString(h, parts[0]+"."+parts[1])
	r, s, err := ecdsa.Sign(rand.Reader, signKey, h.Sum(nil))
	if err != nil {
		cx.Errorf("ecdsa.Sign %v", err)
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}
	sig := base64.URLEncoding.EncodeToString(r.Bytes()) + "." + base64.URLEncoding.EncodeToString(s.Bytes())
	parts[2] = base64.URLEncoding.EncodeToString([]byte(sig))

	replyJSON(w, &ATOKJson{"abelana#accessToken", strings.Join(parts, ".")})
}

// GetSecretKey will send our key in a way that we should only be called once.
func GetSecretKey(w http.ResponseWriter) {
	st := &Status{"abelana#status", base64.URLEncoding.EncodeToString([]byte(abelanaConfig().ServerKey))}
	replyJSON(w, st)
}

/**
 * Access Tokens -- IMPORTANT - This code is here to give us the ability to use Access Tokens before
 * this functality is available in the Google Idenity Toolkit as a standard feature.
 * Once AT's become standard we will switch use them and void our code.
 **/

// AccToken is what we pass to our client, would rather not have the password here as it will
// go away when Idenitty Toolkit supports access tokens.
type AccToken struct {
	UserID string
	Iat    int64
	Exp    int64
}

// Access lets us know if we need another
type Access interface {
	Expired() bool
	ID() string
}

// Expired tells us if we have a valid AuthToken
func (at *AccToken) Expired() bool {
	return time.Now().UTC().After(time.Unix(at.Exp, 0))
}

// ID accessor func for UserID
func (at *AccToken) ID() string {
	return at.UserID
}

// Aauth validates a given AccessToken
func Aauth(c martini.Context, cx appengine.Context, p martini.Params, w http.ResponseWriter) {
	var at *AccToken

	if abelanaConfig().EnableBackdoor && strings.HasPrefix(p["atok"], "LES") {
		at = &AccToken{"00001", time.Now().UTC().Unix(),
			time.Now().UTC().Add(120 * 24 * time.Hour).Unix()}
	} else {
		part := strings.Split(p["atok"], ".")
		if len(part) != 3 {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		h, err := base64.URLEncoding.DecodeString(part[0])
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		hh := struct {
			KeyID string `json:"kid"`
		}{}
		if err = json.Unmarshal(h, &hh); err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		if hh.KeyID != "abelana" {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		ct, err := base64.URLEncoding.DecodeString(part[1])
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		at = &AccToken{}
		if err = json.Unmarshal(ct, &at); err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		if at.UserID == "" || at.Iat == 0 || at.Exp == 0 {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		// Check the signature.
		sig, err := base64.URLEncoding.DecodeString(part[2])
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		hash := md5.New()
		io.WriteString(hash, part[0]+"."+part[1])
		p := strings.Split(string(sig), ".")
		rp, err := base64.URLEncoding.DecodeString(p[0])
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		sp, err := base64.URLEncoding.DecodeString(p[1])
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		r := big.NewInt(0)
		s := big.NewInt(0)
		verify := ecdsa.Verify(&signKey.PublicKey, hash.Sum(nil), r.SetBytes(rp), s.SetBytes(sp))
		if !verify {
			cx.Errorf("CheckSignature %v %v", at.UserID, err)
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
	}

	c.MapTo(at, (*Access)(nil))
}

// decodeSegment decodes the Base64 encoding segment of the JWT token.
// It pads the string if necessary.
func decodeSegment(s string) ([]byte, error) {
	switch len(s) % 4 {
	case 2:
		s = s + "=="
	case 3:
		s = s + "="
	}
	return base64.URLEncoding.DecodeString(s)
}
