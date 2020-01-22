package connection

import (
	"time"

	"github.com/gorilla/websocket"
)

const closeWaitPeriod = time.Second * 5

// CloseConnection closes a websocket connection immediately after sending the specified message
func CloseConnection(wsconn *websocket.Conn, wscode WSCode, message string) {
	wsconn.WriteMessage(WSMTText, []byte(message))

	time.Sleep(closeWaitPeriod)
	wsconn.Close()
}
