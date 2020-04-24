package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/game"
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

	// Game server
	gameServer := game.NewServer()
	gameServer.Init()

	routes.SetupGameServer(&gameServer)

	// for i := 0; i < 100; i++ {
	// 	cards, dtv := game.Generate()
	// 	initialisedCards := game.Initialise(cards, dtv)

	// 	log.Print(cards)
	// 	log.Print(initialisedCards)
	// }

	// Serve
	log.Printf("Game server listening on: %v", address)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
