// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package protocol provides utilities for handling websocket messages.
package protocol

import (
	"encoding/json"
)

// Payload is a wrapper for the payload of a websocket message.
type Payload struct {
	Code    B2Code `json:"code"`
	Message string `json:"message"`
}

// NewPayloadFromBytes tries to create a Payload from the bytes of a websocket message.
// Errors are silent, and will result in a default payload struct being returned.
func NewPayloadFromBytes(payloadBytes []byte) Payload {

	// Attempt to unmarshal the payload bytes into a payload struct. Ignore errors, and
	// just let the payload be returned as is (with default values).
	payload := Payload{}
	_ = json.Unmarshal(payloadBytes, &payload)

	return payload
}
