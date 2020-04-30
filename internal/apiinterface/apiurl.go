package apiinterface

import "fmt"

const apiURL = "https://rahvicrr4g.execute-api.eu-west-2.amazonaws.com"
const apiVersion = "v1"
const endpointProfiles = "profiles"

// GetURL returns the url for the blade2 API
func GetURL() string {
	return fmt.Sprintf("%s/%s/", apiURL, apiVersion)
}
