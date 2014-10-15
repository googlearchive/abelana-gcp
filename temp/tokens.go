// Package token is a set of utilities to validate our GitKit and Access Tokens.  For now, we are
// providing our own Access Tokens, later, we will use GitKit's tokens when they become available.
package token

// "crypto/hmac"
// "crypto/rand"
// "crypto/sha1"
// "encoding/base64"
// "encoding/json"
// "github.com/google/identity-toolkit-go-client/gitkit"

// GetAccessToken validates the GitToken, returns information about the user, and an AccessToken
func GetAccessToken(gittok string) (*User, string, error) {

	return nil, nil, nil
}

// ValidateAccessToken makes sure it's still good.
func ValidateAccessToken(atok string) error {

	return nil
}

// RefreshAccessToken refreshes the access token for another few weeks.
func RefreshAccessToken(atok string) (string, error) {

	return "003 token", nil
}
