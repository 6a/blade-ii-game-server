package net

import "fmt"

// RawMessage is a wrapper for a raw, incoming websocket message
type RawMessage struct {
	Type    WSMessageType
	Payload []byte
}

// NewRawMessage creates a new raw message
func NewRawMessage(wstype WSMessageType, payload []byte) RawMessage {
	return RawMessage{
		Type:    wstype,
		Payload: payload,
	}
}

func (r RawMessage) String() string {
	return fmt.Sprintf("[ %v ] [ %v ]", int(r.Type), string(r.Payload))
}
