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

// waitForMessageAsync asynchronously waits for a websocket to receive (messageCount) number
// of messages, reading them into a channel of the same size. The channel is returned immediately,
// and will be filled when messages are received.
func waitForMessageAsync(wsconn *websocket.Conn, messageCount uint64) chan protocol.Message {

	// Initialize a new channel of the specified size.
	channel := make(chan protocol.Message, messageCount)

	// Start a new goroutine to read from the websocket, as this is a blocking operation.
	go func() {

		// Iterate until the required number of messages have been read.
		for i := uint64(0); i < messageCount; i++ {

			// Block until a new message is received, or the websocket errors.
			mt, payload, err := wsconn.ReadMessage()
			if err != nil {

				// If there was an error, discard the connection.
				Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownConnectionError, err.Error()))
				return
			}

			// If a message was received, package it and add it to the channel.
			messagePayload := protocol.NewPayloadFromBytes(payload)
			packagedMessage := protocol.NewMessageFromPayload(protocol.Type(mt), messagePayload)

			channel <- packagedMessage
		}

	}()

	// Immediately return the channel.
	return channel
}
