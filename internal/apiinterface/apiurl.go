// Copyright 2009 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package apiinterface provides utilities for interacting with the Blade II Online REST API.
package apiinterface

import "fmt"

// apiURL is the URL of the Blade II Online REST API.
const apiURL = "https://rahvicrr4g.execute-api.eu-west-2.amazonaws.com"

// apiVersion is the target Amazon API Gateway stage that should be accessed.
const apiVersion = "v1"

// endpointProfiles is the path of the profiles endpoint of the Blade II Online REST API.
const endpointProfiles = "profiles"

// GetURL constructs and returns the URL for the specified endpoint of the Blade II Online REST API.
func GetURL(endpoint string) string {

	// Concatenate and return the API URL, version, and endpoint path.
	return fmt.Sprintf("%s/%s/%s", apiURL, apiVersion, endpoint)
}
