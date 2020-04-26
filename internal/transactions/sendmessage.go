package transactions

import (
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

func sendMessage(wsconn *websocket.Conn, message protocol.Message) {
	wsconn.WriteMessage(protocol.WSMTText, message.GetPayloadBytes())
}
