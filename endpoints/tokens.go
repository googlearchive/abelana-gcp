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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"appengine"

	"github.com/go-martini/martini"
	"github.com/google/identity-toolkit-go-client/gitkit"
)

////////////////////////////////////////////////////////////////////
const enableBackdoor = true // FIXME(lesv) TEMPORARY BACKDOOR ACCESS
////////////////////////////////////////////////////////////////////

var gclient *gitkit.Client
var serverKey []byte
var publicCerts []*x509.Certificate

// ATOKJson is the json message for an Access Token
type ATOKJson struct {
	Kind string `json:"kind"`
	Atok string `json:"atok"`
}

// tokenInit will setup to use GitKit
func tokenInit() {
	// Provide configuration. gitkit.LoadConfig() can also be used to load
	// the configuration from a JSON file.
	config, err := gitkit.LoadConfig("private/gitkit-server-config.json")
	if err != nil {
		panic("Unable to initialize gitkit config ")
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

	haveCerts(cx)
	dn, err := decodeSegment(p["displayName"])
	pu, err := decodeSegment(p["photoUrl"])
	dName := string(dn)
	photoURL := string(pu)
	if enableBackdoor && p["gittok"] == "Les" {
		err = nil
		token = &gitkit.Token{"Magic", "**AUDIENCE**", time.Now().UTC(),
			time.Now().UTC().Add(1 * time.Hour), "00001", "lesv@abelana-app.com",
			true, "abelana-app.com", "LES001"}
		dName = "Les Vogel"
		photoURL = "https://lh4.googleusercontent.com/-Nt9PfYHmQeI/AAAAAAAAAAI/AAAAAAAAANI/2mbohwDXFKI/photo.jpg?sz=50"
	} else {
		token, err = VerifyToken(p["gittok"]) // TODO FIXME should be gitKit.ValidateToken
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		// TODO verify the Audience is correct
	}

	at := &AccToken{token.LocalID, string(serverKey), time.Now().UTC().Unix(),
		time.Now().UTC().Add(120 * 24 * time.Hour).Unix(), token.Email}

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

	// TODO - Look us up in datastore, add us to memcache, and be happy.
	// We may also want to create us in Datastore
	user, err := FindUser(cx, at.UserID)
	if err != nil {
		// Not found, must create
		user = &User{at.UserID, "firstName", "lastName", dName, token.Email, make([]string, 0, 100)}
		CreateUser(cx, user)
		CopyUserPhoto(cx, photoURL, at.UserID)
	}
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
	HalfPW string // TODO FIXME
	Iat    int64
	Exp    int64
	Email  string
}

// Access lets us know if we need another
type Access interface {
	Expired() bool
	Mail() string
	ID() string
}

// Expired tells us if we have a valid AuthToken
func (at *AccToken) Expired() bool {
	return time.Now().UTC().After(time.Unix(at.Exp, 0))
}

// Mail access func for Email
func (at *AccToken) Mail() string {
	return at.Email
}

// ID accessor func for UserID
func (at *AccToken) ID() string {
	return at.UserID
}

// AtokAuth validates a given AccessToken
func AtokAuth(c martini.Context, cx appengine.Context, p martini.Params, w http.ResponseWriter) {
	var at *AccToken

	haveCerts(cx)
	if enableBackdoor && strings.HasPrefix(p["atok"], "LES") { // FIXME -- TEMPORARY BACKDOOR
		at = &AccToken{"00001", string(serverKey), time.Now().UTC().Unix(),
			time.Now().UTC().Add(120 * 24 * time.Hour).Unix(), "lesv@abelana-app.com"}
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
		if at.UserID == "" || at.HalfPW == "" || at.Iat == 0 || at.Exp == 0 || at.Email == "" {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		// Check the signature.
		s, err := base64.URLEncoding.DecodeString(part[2])
		if err != nil {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		err = publicCerts[0].CheckSignature(x509.SHA256WithRSA, []byte(part[0]+"."+part[1]), s)
		if err != nil {
			cx.Errorf("CheckSignature %v %v", 0, err)
		}

	}

	c.MapTo(at, (*Access)(nil))
}

// RefreshAccessToken refreshes the access token for another few weeks.
func RefreshAccessToken(atok string) (string, error) {

	return "003 token", nil
}

// The following code is modified and taken from the GitKit client library.
// TODO FIXME - I'm turning off Certificate validation for a while

// VerifyToken verifies the JWT is valid and signed by identitytoolkit service
// and returns the verfied token.
func VerifyToken(token string) (*gitkit.Token, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("not a JWT: %s", token)
	}
	// Check the header to extract the "kid" field.
	h, err := decodeSegment(parts[0])
	if err != nil {
		return nil, err
	}
	hh := struct {
		KeyID string `json:"kid"`
	}{}
	if err = json.Unmarshal(h, &hh); err != nil {
		return nil, err
	}
	// cert, err := certs.Cert(hh.KeyID)
	// if err != nil {
	// 	return nil, err
	// }
	// Check the claim set.
	c, err := decodeSegment(parts[1])
	if err != nil {
		return nil, err
	}
	t := struct {
		Iss        string `json:"iss,omitempty"`
		Aud        string `json:"aud,omitempty"`
		Iat        int64  `json:"iat,omitempty"`
		Exp        int64  `json:"exp,omitempty"`
		UserID     string `json:"user_id,omitempty"`
		Email      string `json:"email,omitempty"`
		Verified   bool   `json:"verified,omitempty"`
		ProviderID string `json:"providerId,omitempty"`
	}{}
	if err = json.Unmarshal(c, &t); err != nil {
		return nil, err
	}
	if t.Iss == "" || t.Aud == "" || t.Iat == 0 || t.Exp == 0 || t.UserID == "" {
		return nil, fmt.Errorf("missing required fields: %v", t)
	}
	// Check the signature.
	// s, err := decodeSegment(parts[2])
	// if err != nil {
	// 	return nil, err
	// }
	// FIXME -- turn off cert validation
	// if err := cert.CheckSignature(x509.SHA256WithRSA, []byte(parts[0]+"."+parts[1]), s); err != nil {
	// 	return nil, err
	// }
	return &gitkit.Token{
		Issuer:        t.Iss,
		Audience:      t.Aud,
		IssueAt:       time.Unix(t.Iat, 0),
		ExpireAt:      time.Unix(t.Exp, 0),
		LocalID:       t.UserID,
		Email:         t.Email,
		EmailVerified: t.Verified,
		ProviderID:    t.ProviderID,
		TokenString:   token,
	}, nil
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
