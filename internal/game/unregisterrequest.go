// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game provides implements the Blade II Online game server.
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
