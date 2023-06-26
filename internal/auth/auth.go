package auth

import (
	"errors"
	"net/http"
	"strings"
)

// GetAPIKey
// GetAPIKey returns the API key from the request headers
// Example:
// Authorization: ApiKey {API_KEY}
func GetAPIKey(headers http.Header) (string, error) {
	authHeaderVal := headers.Get("Authorization");
	if authHeaderVal == "" {
		return "", errors.New("no authentication info found")
	}

	vals := strings.Split(authHeaderVal, " ")
	if len(vals) != 2 {
		return "", errors.New("malformed authorization header")
	}

	if vals[0] != "ApiKey" {
		return "", errors.New("malformed first part of authorization header")
	}

	return vals[1], nil
}