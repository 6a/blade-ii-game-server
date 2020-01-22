package routes

import (
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/gatekeeper"
	"github.com/6a/blade-ii-game-server/internal/matchmaking"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // TODO replace with a proper method
}

// SetupMatchMaking sets up the matchmaking queue endpoint
func SetupMatchMaking(mm *matchmaking.MatchMaking) {
	http.HandleFunc("/matchmaking", func(w http.ResponseWriter, r *http.Request) {
		wsconn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// Close conn
		}

		go gatekeeper.Handle(wsconn, mm)
	})
}
