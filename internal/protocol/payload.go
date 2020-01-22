package protocol

import "encoding/json"

// Payload is a wrapper for the payload of a websocket message
type Payload struct {
	Code    B2Code `json:"code"`
	Message string `json:"message"`
}

// NewPayloadFromBytes tries to create a Payload from the bytes of a websocket message
func NewPayloadFromBytes(payloadBytes []byte) Payload {
	payload := Payload{}
	_ = json.Unmarshal(payloadBytes, &payload)

	return payload
}
