// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package routes defines http endpoint handlers for http/websocket connections to the server.
package routes

import (
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/game"
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/6a/blade-ii-game-server/internal/transactions"
)

// SetupGameServer sets up the game server endpoint. Pass in a pointer to the game server.
func SetupGameServer(gs *game.Server) {

	// Defines the handler for the /game endpoint.
	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {

		// On connection, upgrade the connection to a websocket connection.
		wsconn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {

			// if errored, discard the connection early.
			transactions.Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthBadCredentials, err.Error()))
		}

		// If the upgrade was successful, pass connection and the game server pointer to another handler (using a goroutine to
		// avoid blocking) which will perform authentication and match validity checking, and handle adding the client to the
		// game server.
		go transactions.HandleGSConnection(wsconn, gs)
	})
}
