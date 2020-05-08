// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package transactions implements handlers for various interactions with raw websocket connections,
// before they are packaged and added to the server.
package transactions

import (
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

// sendMessage asynchronously sends a message down the websocket.
func sendMessage(wsconn *websocket.Conn, message protocol.Message) {

	// Write the message, ignoring any errors.
	wsconn.WriteMessage(protocol.WSMTText, message.GetPayloadBytes())
}
