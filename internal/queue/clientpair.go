package queue

import (
	"fmt"
	"time"

	"github.com/6a/blade-ii-game-server/internal/protocol"
)

// ClientPair is a light wrapper for a pair of client connections
type ClientPair struct {
	C1              *MMClient
	C2              *MMClient
	ReadyStart      time.Time
	IsReadyChecking bool
}

// NewPair creates a new ClientPair
func NewPair(c1 *MMClient, c2 *MMClient) ClientPair {
	return ClientPair{
		C1: c1,
		C2: c2,
	}
}

// SendMatchStartMessage sends a match start message to both clients
func (pair *ClientPair) SendMatchStartMessage() {
	pair.ReadyStart = time.Now()
	pair.IsReadyChecking = true

	pair.C1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchMakingGameFound, ""))
	pair.C1.IsReadyChecking = true

	pair.C2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchMakingGameFound, ""))
	pair.C2.IsReadyChecking = true
}

// SendMatchConfirmedMessage sends a match confirmation message with match ID to both clients
func (pair *ClientPair) SendMatchConfirmedMessage(matchID int64) {
	stringID := fmt.Sprintf("%v", matchID)
	pair.C1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchConfirmed, stringID))
	pair.C2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchConfirmed, stringID))
}
