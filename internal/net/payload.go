package net

// Payload is a wrapper for the payload of a websocket message
type Payload struct {
	Code    WSCode `json:"code"`
	Message string `json:"message"`
}
