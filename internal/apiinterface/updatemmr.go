// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package apiinterface provides utilities for interacting with the Blade II Online REST API.
package apiinterface

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// UpdateMatchStats synchronously sends a request to the API server to update the MMR, as well as
// the w/d/l for the specified players, based on the winner.
//
// Fails silently (for the client) but logs to console.
func UpdateMatchStats(client1ID uint64, client2ID uint64, winner Winner) {

	// Create an instance of the match update request struct, with the parameters that were passed in.
	updateRequest := MMRUpdateRequest{
		client1ID,
		client2ID,
		winner,
	}

	// Create a JSON formatting string based on the match update request.
	updateRequestBytes, err := json.Marshal(updateRequest)
	if err != nil {
		log.Printf("Error packaging MMR update data: %v", err.Error())
		return
	}

	// Create a temporary instance of a http client.
	var client http.Client

	// Set up the request that will be sent to the API.
	req, err := http.NewRequest(http.MethodPatch, GetURL(endpointProfiles), bytes.NewBuffer(updateRequestBytes))
	if err != nil {
		log.Printf("Error packaging MMR update data: %v", err.Error())
		return
	}

	// Add required auth header to the request.
	addAuthHeader(req)

	// Attempt to make the request that was set up above.
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error Sending MMR update: %s", err.Error())
	} else if resp.StatusCode != http.StatusNoContent {

		// Defer the closing of the response body stream so that it will be cleaned up properly when this closure is exited.
		defer resp.Body.Close()

		// Attempt to read the contents of the response body, and try to determine what the error was.
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error sending MMR update: %v", err.Error())
		} else {
			log.Printf("Error Sending MMR update: %v", string(body))
		}
	} else {
		log.Println("Successfully updated Match stats")
	}
}
