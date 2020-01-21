package net

// ConnectAuth is the message body for connection requests
type ConnectAuth struct {
	UID string `json:"uid"`
	Key string `json:"key"`
}
