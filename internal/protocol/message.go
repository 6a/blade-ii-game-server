// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package protocol provides utilities for handling websocket messages.
package protocol

import "encoding/json"

// Message is a wrapper for an outgoing websocket message and its message type.
type Message struct {
	Type    Type
	Payload Payload
}

// NewMessage creates and returns new message.
func NewMessage(wstype Type, instructionCode B2Code, payload string) Message {
	return Message{
		Type: wstype,
		Payload: Payload{
			Code:    instructionCode,
			Message: payload,
		},
	}
}

// NewMessageFromPayload creates and returns new message, with the specified payload.
func NewMessageFromPayload(wstype Type, payload Payload) Message {
	return Message{
		Type:    wstype,
		Payload: payload,
	}
}

// GetPayloadBytes returns the payload of the message as a byte array.
func (r Message) GetPayloadBytes() []byte {

	// Marshal the payload object into a json string. Errors are ignored.
	jsonString, _ := json.Marshal(r.Payload)

	// Convert the json string to a byte array before returning.
	return []byte(jsonString)
}
