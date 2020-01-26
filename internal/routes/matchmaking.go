package routes

import (
	"net/http"

	"github.com/6a/blade-ii-matchmaking-server/internal/protocol"
	"github.com/6a/blade-ii-matchmaking-server/internal/transactions"

	"github.com/6a/blade-ii-matchmaking-server/internal/matchmaking"
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
			transactions.Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthBadCredentials, err.Error()))
		}

		go transactions.HandleMMConnection(wsconn, mm)
	})
}
