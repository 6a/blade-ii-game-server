package apiinterface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// UpdateMatchStats synchronously sends a request to the API server to update the mmr, as well as the w/d/l for the specified players
// based on the winner
// Fails silently (for the client) but logs to console
func UpdateMatchStats(client1ID uint64, client2ID uint64, winner Winner) {
	updateRequest := MMRUpdateRequest{
		client1ID,
		client2ID,
		winner,
	}

	updateRequestBytes, err := json.Marshal(updateRequest)
	if err != nil {
		log.Printf("Error packaging MMR update data: %v", err.Error())
		return
	}

	var client http.Client

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s%s", GetURL(), endpointProfiles), bytes.NewBuffer(updateRequestBytes))
	if err != nil {
		log.Printf("Error packaging MMR update data: %v", err.Error())
		return
	}

	addAuthHeader(req)

	resp, err := client.Do(req)

	if resp.StatusCode != http.StatusNoContent {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			log.Printf("Error Sending MMR update: %v", string(body))
		} else {
			log.Print("Error Sending MMR update: unknown error")
		}
	} else {
		log.Print("Successfully updated Match stats")
	}
}
