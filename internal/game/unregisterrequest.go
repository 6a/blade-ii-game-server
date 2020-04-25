package game

import (
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

// UnregisterRequest is a wrapper for the information required to remove a client from the queue
type UnregisterRequest struct {
	Client  *GClient
	Reason  protocol.B2Code
	Message string
}
