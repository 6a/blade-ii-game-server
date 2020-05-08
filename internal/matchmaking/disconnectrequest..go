// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package matchmaking implements the Blade II Online matchmaking server.
package matchmaking

import (
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

// DisconnectRequest is a wrapper for the information required to remove a client from the matchmaking queue.
type DisconnectRequest struct {

	// The internal matchmaking queue ID for this client.
	clientIndex uint64

	// The reason for removal.
	Reason protocol.B2Code

	// An optional message, to be sent to the client before disconnecting.
	Message string
}
