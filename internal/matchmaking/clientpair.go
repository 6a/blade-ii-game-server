// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package matchmaking implements the Blade II Online matchmaking server.
package matchmaking

import (
	"strconv"
	"time"

	"github.com/6a/blade-ii-game-server/internal/protocol"
)

// ClientPair is a light wrapper for a pair of client connections.
type ClientPair struct {

	// Pointer to both clients.
	Client1 *MMClient
	Client2 *MMClient

	// The time at which the ready check began.
	ReadyStart time.Time

	// Whether the pair is currently undergoing a ready check.
	IsReadyChecking bool
}

// NewPair initializes and returns a pointer to a new client pair.
func NewPair(client1 *MMClient, client2 *MMClient) *ClientPair {

	// Initialize and return a new clientpair based on the specified clients.
	return &ClientPair{
		Client1: client1,
		Client2: client2,
	}
}

// SendMatchFoundMessage sends a match found message to both clients.
func (pair *ClientPair) SendMatchFoundMessage() {

	// Update the state of the pair.
	pair.ReadyStart = time.Now()
	pair.IsReadyChecking = true

	// Send a match found message to client 1, and set their internal ready checking flag to true.
	pair.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchMakingMatchFound, ""))
	pair.Client1.IsReadyChecking = true

	// Send a match found message to client 2, and set their internal ready checking flag to true.
	pair.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchMakingMatchFound, ""))
	pair.Client2.IsReadyChecking = true
}

// SendMatchConfirmedMessage sends a match confirmation message with match ID to both clients.
func (pair *ClientPair) SendMatchConfirmedMessage(matchID uint64) {

	// Get a string representation of the match ID.
	matchIDString := strconv.FormatUint(matchID, 10)

	// Send the match ID string to both clients.
	pair.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchConfirmed, matchIDString))
	pair.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchConfirmed, matchIDString))
}
