package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/matchmaking"
	"github.com/6a/blade-ii-game-server/internal/routes"
)

const address = "localhost:80"
const addressFlag = "address"

var addr = flag.String(addressFlag, address, "Service Address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	// Init database connection
	database.Init()

	// Matchmaking server
	matchMakingServer := matchmaking.NewMatchMaking()
	matchMakingServer.Init()

	routes.SetupMatchMaking(&matchMakingServer)

	// Games

	// Serve
	log.Printf("Matchmaking server listening on: %v", address)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
