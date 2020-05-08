// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package routes defines http endpoint handlers for http/websocket connections to the server.
package routes

import (
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/matchmaking"
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/6a/blade-ii-game-server/internal/transactions"
)

// SetupMatchMaking sets up the matchmaking server endpoint. Pass in a pointer to the matchmaking server.
func SetupMatchMaking(mm *matchmaking.Server) {

	// Defines the handler for the /matchmaking endpoint.
	http.HandleFunc("/matchmaking", func(w http.ResponseWriter, r *http.Request) {

		// On connection, upgrade the connection to a websocket connection.
		wsconn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {

			// if errored, discard the connection early.
			transactions.Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthBadCredentials, err.Error()))
		}

		// If the upgrade was successful, pass connection and the matchmaking server pointer to another handler (using a goroutine to
		// avoid blocking) which will perform authentication, and handle adding the client to the matchmaking queue.
		go transactions.HandleMMConnection(wsconn, mm)
	})
}
