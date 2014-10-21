// Package token is a set of utilities to validate our GitKit and Access Tokens.  For now, we are
// providing our own Access Tokens, later, we will use GitKit's tokens when they become available.
package abelanaEndpoints

// "crypto/hmac"
// "crypto/rand"
// "crypto/sha1"
// "encoding/base64"
// "encoding/json"
// "github.com/google/identity-toolkit-go-client/gitkit"

import (
	"encoding/base64"
	"encoding/json"
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
var publicCerts []appengine.Certificate

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
		log.Print("unable to read serverKey")
		panic("Unable to read serverKey")
	}
	log.Printf("serverKey: %v", string(serverKey))
}

// haveCerts - make sure we have the certificates.
func haveCerts(c appengine.Context) {
	if len(publicCerts) > 0 {
		return
	}
	publicCerts, err := appengine.PublicCertificates(c)
	if err != nil {
		panic("unable to get certs")
	}
	c.Debugf("haveCerts: %v %v", len(publicCerts), publicCerts[0].KeyName)
}

type TokenResponse struct {
	Status string
	ATok   string
}

// GitAuth - see if the token is valid
func Login(c appengine.Context, p martini.Params, w http.ResponseWriter) {
	var token *gitkit.Token
	var err error

	if enableBackdoor && p["gittok"] == "Les" {
		err = nil
		token = &gitkit.Token{"Magic", "**AUDIENCE**", time.Now().UTC(),
			time.Now().UTC().Add(1 * time.Hour), "00001", "lesv@angeltech.com",
			true, "abelana-app.com", "LES001"}
	} else {
		token, err = gclient.ValidateToken(p["gittok"])
		if err != nil {
			log.Printf("unable to validate token %v", err)
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}
		// TODO verify the Audience is correct
	}
	c.Debugf("Login.token %v", token)

	accToken := &AccToken{token.LocalID, string(serverKey), time.Now().UTC(),
		time.Now().UTC().Add(120 * 24 * time.Hour), token.Email}

	parts := make([]string, 3)

	parts[0] = base64.URLEncoding.EncodeToString([]byte(`{"kid": "abelana"}`))
	tok, err := json.Marshal(accToken)
	if err != nil {
		c.Errorf("Marshal JSON %v", err)
	}
	parts[1] = base64.URLEncoding.EncodeToString(tok)
	str, sig, err := appengine.SignBytes(c, []byte(parts[0]+"."+parts[1]))
	if err != nil {
		c.Errorf("Login.Sign %v", err)
		return
	}
	parts[2] = base64.URLEncoding.EncodeToString(sig)
	log.Printf("Login.sign keyname? %v", str)

	replyJson(w, &TokenResponse{"Ok", strings.Join(parts, ".")})
}

/**
 * Access Tokens -- IMPORTANT - This code is here to give us the ability to use Access Tokens before
 * this functality is available in the Google Idenity Toolkit as a standard feature.
 * Once AT's become standard we will switch use them and void our code.
 **/
type AccToken struct {
	UserID string
	HalfPW string
	Iat    time.Time
	Exp    time.Time
	Email  string
}

// Expires tells us if we have a valid AuthToken
func (at *AccToken) Expired() bool {
	return time.Now().UTC().After(at.Exp)
}

// AtokAuth validates a given AccessToken
func AtokAuth(p martini.Params, w http.ResponseWriter) {
	if !strings.HasPrefix(p["atok"], "LES") { // FIXME -- TEMPORARY BACKDOOR
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}
}

// GetAccessToken validates the GitToken, returns information about the user, and an AccessToken
// func GetAccessToken(gittok string) (*User, string, error) {
//
// 	return nil, nil, nil
// }

// ValidateAccessToken makes sure it's still good.
func ValidateAccessToken(atok string) error {

	return nil
}

// RefreshAccessToken refreshes the access token for another few weeks.
func RefreshAccessToken(atok string) (string, error) {

	return "003 token", nil
}
