// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package apiinterface provides utilities for interacting with the Blade II Online REST API.
package apiinterface

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
)

// Read the auth details for the API from the environment variables.
var (
	apiUsername = os.Getenv("api_username")
	apiPassword = os.Getenv("api_password")
)

// addAuthHeader adds a 'Basic' HTTP Authentication Scheme header (RFC7617) to the specified request.
func addAuthHeader(req *http.Request) {

	// Create the credentials string, and encode it using standard base64 encoding as per the spec (RFC7617).
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(apiUsername + ":" + apiPassword))

	// Combine the appropriate prefix with the encoded credentials.
	authHeaderString := fmt.Sprintf("Basic %s", encodedCredentials)

	// Add the 'Authorization' header to the existing collection of headers, in the format specified by the spec (RFC7617).
	req.Header.Add("Authorization", authHeaderString)
}
