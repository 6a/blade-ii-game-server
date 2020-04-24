package routes

import (
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/6a/blade-ii-game-server/internal/transactions"

	"github.com/6a/blade-ii-game-server/internal/matchmaking"
	"github.com/gorilla/websocket"
)

var mmupgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // TODO replace with a proper method
}

// SetupMatchMaking sets up the matchmaking queue endpoint
func SetupMatchMaking(mm *matchmaking.Server) {
	http.HandleFunc("/matchmaking", func(w http.ResponseWriter, r *http.Request) {
		wsconn, err := mmupgrader.Upgrade(w, r, nil)
		if err != nil {
			transactions.Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthBadCredentials, err.Error()))
		}

		go transactions.HandleMMConnection(wsconn, mm)
	})
}
