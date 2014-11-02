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
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"appengine"

	"github.com/go-martini/martini"
	"github.com/google/identity-toolkit-go-client/gitkit"
)

var (
	gclient     *gitkit.Client
	serverKey   []byte
	publicCerts []*x509.Certificate
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
		log.Fatalf("Unable to initialize gitkit config")
	}
	gclient, err = gitkit.New(config)
	if err != nil {
		log.Printf("new Client ** %v", err)
		panic("unable to init gitkit")
	}
	serverKey, err = ioutil.ReadFile("private/serverpw")
	if err != nil {
		panic("Unable to read serverKey")
	}
}

// haveCerts - make sure we have the certificates.
func haveCerts(cx appengine.Context) {
	if len(publicCerts) > 0 {
		return
	}
	certs, err := appengine.PublicCertificates(cx)
	if err != nil {
		panic("unable to get certs")
	}

	publicCerts = make([]*x509.Certificate, len(certs))
	for i, cert := range certs {
		block, _ := pem.Decode([]byte(cert.Data))
		if block == nil {
			panic("failed to parse certificate PEM")
		}
		publicCerts[i], err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			panic("failed to parse certificate: " + err.Error())
		}
	}
}

// Login - see if the token is valid
func Login(cx appengine.Context, p martini.Params, w http.ResponseWriter) {
	var token *gitkit.Token
	var err error
	var dName, photoURL string

	haveCerts(cx)
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
	_, sig, err := appengine.SignBytes(cx, []byte(parts[0]+"."+parts[1]))
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}
	parts[2] = base64.URLEncoding.EncodeToString(sig)

	replyJSON(w, &ATOKJson{"abelana#accessToken", strings.Join(parts, ".")})

	// Look us up in datastore and be happy.
	_, err = findUser(cx, at.UserID)
	if err != nil {
		// Not found, must create
		createUser(cx, User{UserID: at.UserID, DisplayName: dName, Email: token.Email})
		if photoURL != "" {
			delayCopyUserPhoto.Call(cx, photoURL, at.UserID)
		}
	}
}

// Refresh will refresh an Access Token (ATok)
func Refresh(cx appengine.Context, p martini.Params, w http.ResponseWriter) {
	haveCerts(cx)
	s := strings.Split(p["atok"], ".")
	ct, err := base64.URLEncoding.DecodeString(s[1])
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
	s[1] = base64.URLEncoding.EncodeToString(ts)
	_, sig, err := appengine.SignBytes(cx, []byte(s[0]+"."+s[1]))
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}
	s[2] = base64.URLEncoding.EncodeToString(sig)

	replyJSON(w, &ATOKJson{"abelana#accessToken", strings.Join(s, ".")})
}

// GetSecretKey will send our key in a way that we only need call this once.
func GetSecretKey(w http.ResponseWriter) {
	st := &Status{"abelana#status", base64.URLEncoding.EncodeToString(serverKey)}
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

	haveCerts(cx)
	// FIXME -- TEMPORARY BACKDOOR
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
		s, err := base64.URLEncoding.DecodeString(part[2])
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		for _, cert := range publicCerts {
			err = cert.CheckSignature(x509.SHA256WithRSA, []byte(part[0]+"."+part[1]), s)
			if err == nil {
				break
			}
		}
		if err != nil {
			cx.Errorf("CheckSignature %v %v", at.UserID, err)
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
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
