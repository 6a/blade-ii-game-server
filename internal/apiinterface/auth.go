package apiinterface

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
)

var apiUsername = os.Getenv("api_username")
var apiPassword = os.Getenv("api_password")

// addAuthHeader adds the auth header to the specified request
func addAuthHeader(req *http.Request) {
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(apiUsername + ":" + apiPassword))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedAuth))
}
