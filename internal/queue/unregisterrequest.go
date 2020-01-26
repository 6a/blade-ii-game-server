package queue

import (
	"github.com/6a/blade-ii-matchmaking-server/internal/protocol"
)

// UnregisterRequest is a wrapper for the information required to remove a client from the queue
type UnregisterRequest struct {
	clientID uint64
	Reason   protocol.B2Code
	Message  string
}
