package connection

import (
	"time"

	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

const closeWaitPeriod = time.Second * 5

// Close closes a websocket connection immediately after sending the specified message
func Close(wsconn *websocket.Conn, message protocol.Message) {
	wsconn.WriteMessage(protocol.WSMTText, message.GetPayloadBytes())

	time.Sleep(closeWaitPeriod)
	wsconn.Close()
}
