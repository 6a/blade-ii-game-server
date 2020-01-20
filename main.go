package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/routes"
)

const (
	certPath = "crypto/server.crt"
	keyPath  = "crypto/server.key"
)

var addr = flag.String("wsaddress", "127.0.0.1:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	http.HandleFunc("/queue", routes.Queue)
	log.Fatal(http.ListenAndServeTLS(*addr, certPath, keyPath, nil))
}