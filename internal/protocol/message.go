package protocol

import "encoding/json"

// Message is a wrapper for an outgoing websocket message and its message type
type Message struct {
	Type    Type
	Payload Payload
}

// NewMessage creates a new message
func NewMessage(wstype Type, instructionCode B2Code, payload string) Message {
	return Message{
		Type: wstype,
		Payload: Payload{
			Code:    instructionCode,
			Message: payload,
		},
	}
}

// NewMessageFromPayload creates a new message from a payload object
func NewMessageFromPayload(wstype Type, payload Payload) Message {
	return Message{
		Type:    wstype,
		Payload: payload,
	}
}

// GetPayloadBytes returns the payload of the message as a byte array
func (r Message) GetPayloadBytes() []byte {
	jsonString, _ := json.Marshal(r.Payload)
	return []byte(jsonString)
}
