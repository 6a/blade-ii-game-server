package connection

import "encoding/json"

// Message is a wrapper for an outgoing websocket message and its message type
type Message struct {
	Type    WSMessageType
	Payload Payload
}

// NewMessage creates a new message
func NewMessage(wstype WSMessageType, instructionCode WSCode, payload string) Message {
	return Message{
		Type: wstype,
		Payload: Payload{
			Code:    instructionCode,
			Message: payload,
		},
	}
}

// GetPayloadBytes returns the payload of the message as a byte array
func (r Message) GetPayloadBytes() []byte {
	jsonString, _ := json.Marshal(r.Payload)
	return []byte(jsonString)
}
