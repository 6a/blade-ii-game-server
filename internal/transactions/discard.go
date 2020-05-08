// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package transactions implements handlers for various interactions with raw websocket connections,
// before they are packaged and added to the server.
package transactions

import (
	"time"

	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

// closeWaitPeriod is the amount of time to wait after sending a close message, before closing a websocket.
const closeWaitPeriod = time.Second * 1

// Discard sends the specified message down the websocket, and then closes it after blocking for a small
// period of time.
func Discard(wsconn *websocket.Conn, message protocol.Message) {

	// Write the message to the websocket. Errors are ignored.
	wsconn.WriteMessage(protocol.WSMTText, message.GetPayloadBytes())

	// Sleep for a short time to allow the message to be sent.
	time.Sleep(closeWaitPeriod)

	// Close the websocket.
	wsconn.Close()
}
