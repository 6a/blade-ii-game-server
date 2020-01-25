package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/matchmaking"
	"github.com/6a/blade-ii-game-server/internal/routes"
)

const (
	certPath = "crypto/server.crt"
	keyPath  = "crypto/server.key"
)

const address = "127.0.0.1:8080"

var addr = flag.String("address", address, "Service Address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	// Init database connection
	database.Init()

	// Matchmaking queue
	matchMakingQueue := matchmaking.NewMatchMaking()
	matchMakingQueue.Init()

	routes.SetupMatchMaking(&matchMakingQueue)

	// Games

	// Serve
	log.Printf("Game server listening on: %v", address)
	log.Fatal(http.ListenAndServeTLS(*addr, certPath, keyPath, nil))
}
