package queue

import (
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

type UnregisterRequest struct {
	clientID uint64
	Reason   protocol.B2Code
	Message  string
}
