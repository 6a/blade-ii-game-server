package net

import "fmt"

// RawMessage is a wrapper for a websocket message
type RawMessage struct {
	Type    WSMessageType
	Payload []byte
	Error   error
}

func MakeRawMessage(wstype WSMessageType, payload []byte, err error) RawMessage {
	return RawMessage{
		Type:    wstype,
		Payload: payload,
		Error:   err,
	}
}

func (r RawMessage) String() string {
	return fmt.Sprintf("[ %v ] [ %v ] [ %v ]", int(r.Type), string(r.Payload), r.Error)
}
