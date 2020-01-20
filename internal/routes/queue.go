package routes

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/mm"
	"github.com/6a/blade-ii-game-server/internal/net"
	"github.com/gorilla/websocket" 
)

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // TODO replace with a proper method
}

// Queue is the matchmaking queue endpoint
func Queue(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Close conn
	}

	connectRequest := net.ConnectRequest{}
	err = json.Unmarshal([]byte(body), &connectRequest)
	if err != nil {
		// Close conn
	}

	err = database.ValidateAuth(connectRequest.UID, connectRequest.Key)
	if err != nil {
		// Close conn
	}

	mmr, err := database.GetMMR(connectRequest.UID)
	if err != nil {
		// Close conn
	}

	wsconn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Close conn
	}

	client := net.NewClient(wsconn, connectRequest.UID, mmr)
	mm.JoinQueue(&client) 
}
