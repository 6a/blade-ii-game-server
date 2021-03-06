// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package main implements the initialization and entry point function for this server, in main().
package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/6a/blade-ii-game-server/internal/matchmaking"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/game"
	"github.com/6a/blade-ii-game-server/internal/routes"
)

// address is the local address:port that this server will be available on,
const address = "localhost:20000"

func main() {
	// Seed the random package.
	rand.Seed(time.Now().UTC().UnixNano())

	// Initialise the database package.
	database.Init()

	// Create and initialise an instance of the game server.
	gameServer := game.NewServer()

	// Set up the game server http handler.
	routes.SetupGameServer(gameServer)

	// Create and initialise instance of the matchmaking server.
	matchmakingServer := matchmaking.NewServer()

	// Set up the matchmaking server http handler.
	routes.SetupMatchMaking(matchmakingServer)

	log.Printf("Blade II Online Gameserver listening on: %v", address)

	// Start the http server - the log.Fatal wrapper ensures that any exceptions will cause a clean exit with a proper exit code.
	log.Fatal(http.ListenAndServe(address, nil))
}
