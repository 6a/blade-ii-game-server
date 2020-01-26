package transactions

import (
	"time"

	"github.com/6a/blade-ii-matchmaking-server/internal/protocol"
	"github.com/gorilla/websocket"
)

const closeWaitPeriod = time.Second * 1

// Discard sends a close message and later closes a websocket connection
func Discard(wsconn *websocket.Conn, message protocol.Message) {
	wsconn.WriteMessage(protocol.WSMTText, message.GetPayloadBytes())

	time.Sleep(closeWaitPeriod)
	wsconn.Close()
}
