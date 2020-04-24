package routes

import (
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/game"
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/6a/blade-ii-game-server/internal/transactions"

	"github.com/gorilla/websocket"
)

var gsupgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // TODO replace with a proper method
}

// SetupGameServer sets up the game server endpoint
func SetupGameServer(gs *game.Server) {
	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		wsconn, err := gsupgrader.Upgrade(w, r, nil)
		if err != nil {
			transactions.Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthBadCredentials, err.Error()))
		}

		go transactions.HandleGSConnection(wsconn, gs)
	})
}
