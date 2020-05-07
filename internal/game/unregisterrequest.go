// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game provides implements the Blade II Online game server.
package game

import (
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

// DisconnectRequest is a wrapper for the information required to remove a client from the game server.
type DisconnectRequest struct {

	// A pointer to the client to remove.
	Client *GClient

	// The reason for removal.
	Reason protocol.B2Code

	// An optional to be sent, either to just the player, or both, depending on whether they are
	// in a match, and the state of the match etc..
	Message string
}
