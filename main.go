package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/net"
	"github.com/6a/blade-ii-game-server/internal/routes"
)

const (
	certPath = "crypto/server.crt"
	keyPath  = "crypto/server.key"
)

var addr = flag.String("address", "127.0.0.1:8080", "Service Address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	// Matchmaking queue
	matchMakingQueue := net.NewMatchMaking()
	routes.SetupMatchMaking(&matchMakingQueue)

	// Games

	// Serve
	log.Fatal(http.ListenAndServeTLS(*addr, certPath, keyPath, nil))
}
