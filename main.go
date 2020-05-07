package main

import (
	"log"
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/matchmaking"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/game"
	"github.com/6a/blade-ii-game-server/internal/routes"
)

const address = "localhost:20000"

func main() {
	// Init database connection
	database.Init()

	// Game server
	gameServer := game.NewServer()
	gameServer.Init()

	routes.SetupGameServer(gameServer)

	// Matchmaking server
	matchmakingServer := matchmaking.NewServer()
	matchmakingServer.Init()

	routes.SetupMatchMaking(matchmakingServer)

	// Serve
	log.Printf("Blade II Online Gameserver listening on: %v", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
