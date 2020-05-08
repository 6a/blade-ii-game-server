// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package transactions implements handlers for various interactions with raw websocket connections,
// before they are packaged and added to the server.
package transactions

import (
	"log"
	"time"

	"github.com/6a/blade-ii-game-server/internal/game"
	"github.com/6a/blade-ii-game-server/internal/protocol"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/matchmaking"
	"github.com/gorilla/websocket"
)

// connectionTimeOut is the maximum amount of time to wait for auth and/or match data after a websocket connection
// is made.
const connectionTimeOut = time.Second * 10

// HandleGSConnection waits for the new connection to send an auth protocol.
// Once received, it checks if the auth is valid, and then waits for the
// connection to send a match ID.
//
// If it does not receive an auth message and match ID within the timeout period, it drops the
// connection.
func HandleGSConnection(wsconn *websocket.Conn, gs *game.Server) {

	// Set up an async wait queue, to wait for (2) messages from the websocket
	inChannel := waitForMessageAsync(wsconn, 2)

	// Declare some values that set and/or read during various stages of the connection handler.
	var databaseID uint64
	var publicID string
	var b2ErrorCode protocol.B2Code
	var err error
	var authReceived bool = false

	// Loop until control exits.
	for {

		// Select will block, waiting for channel writes, until the timeout period is reached, where it will then
		// discard the connection and exit.
		select {
		case res := <-inChannel:

			// If auth has not yet been received, this should be the first message. Attempt to authentication using the
			// provided data.
			if !authReceived {

				// Send a message to the client indicating that the auth data was received.
				sendMessage(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthReceived, ""))

				// Validate the credentials in the payload. Errors lead to this function exiting immediately after
				// discarding the websocket connection.
				databaseID, publicID, b2ErrorCode, err = checkAuth(res.Payload)
				if err != nil {
					Discard(wsconn, protocol.NewMessage(protocol.WSMTText, b2ErrorCode, err.Error()))
					return
				}

				// If we reach here, authentication was successfull, and we inform the client accordingly.
				sendMessage(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthSuccess, ""))

				// Also set the auth received flag so that the next message from the client is handled as match
				// data.
				authReceived = true
			} else {

				// Send a message to the client indicating that the match data was received.
				sendMessage(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchIDReceived, ""))

				// Validate the match data. Errors lead to this function exiting immediately after
				// discarding the websocket connection.
				matchID, b2code, err := validateMatch(databaseID, res.Payload)
				if err != nil {
					Discard(wsconn, protocol.NewMessage(protocol.WSMTText, b2code, err.Error()))
					return
				}

				// If we reach here, the match data was confirmed as valid, and we inform the client accordingly.
				sendMessage(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchIDConfirmed, ""))

				// Grab the clients display name and avatar as well - if this errors, log it and use a placeholder.
				displayname, avatar, err := database.GetClientNameAndAvatar(databaseID)
				if err != nil {
					log.Printf("Error getting displayname for user [ %d ]: %s", databaseID, err.Error())
					displayname = "<unknown>"
				}

				// Pass the websocket connection to the game server to package and add.
				gs.AddClient(wsconn, databaseID, publicID, displayname, avatar, matchID)
				return
			}
		case <-time.After(connectionTimeOut):

			// If the connection timed out, discard the connection with an appropriate message.
			if !authReceived {
				Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthNotReceived, "Auth not received"))
			} else {
				Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchIDNotReceived, "Match ID not received"))
			}

			return
		}
	}
}

// HandleMMConnection waits for the new connection to send an auth protocol.
// Once received, it checks if the auth is valid, then retrieves the mmr of the
// specified account.
//
// If it does not receive an auth message within the timeout period, it drops the
// connection.
func HandleMMConnection(wsconn *websocket.Conn, mm *matchmaking.Server) {

	// Set up an async wait queue, to check for 1 message from the websocket.
	authChannel := waitForMessageAsync(wsconn, 1)

	// Select will block, waiting for channel writes, until the timeout period is reached, where it will then
	// discard the connection and exit.
	select {
	case res := <-authChannel:

		// Validate the credentials in the payload. Errors lead to this function exiting immediately after
		// discarding the websocket connection.
		databaseID, publicID, b2ErrorCode, err := checkAuth(res.Payload)
		if err != nil {
			Discard(wsconn, protocol.NewMessage(protocol.WSMTText, b2ErrorCode, err.Error()))
			return
		}

		// Get the MMR for the authenticated player. Errors cause this function to exit immediately after
		// discarding the websocket connection.
		mmr, err := database.GetMMR(databaseID)
		if err != nil {
			Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownConnectionError, err.Error()))
			return
		}

		// Pass the websocket connection to the matchmaking server to package and add.
		mm.AddClient(wsconn, databaseID, publicID, mmr)
	case <-time.After(connectionTimeOut):

		// If the connection timed out, discard the connection with an appropriate message.
		Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownConnectionError, "Auth message not received"))
		return
	}
}
