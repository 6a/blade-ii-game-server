package transactions

import (
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

func waitForMessageAsync(wsconn *websocket.Conn, messageCount uint64) chan protocol.Message {
	channel := make(chan protocol.Message, messageCount)

	go func() {
		for i := uint64(0); i < messageCount; i++ {
			mt, payload, err := wsconn.ReadMessage()
			if err != nil {
				Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownConnectionError, err.Error()))
				return
			}

			messagePayload := protocol.NewPayloadFromBytes(payload)
			packagedMessage := protocol.NewMessageFromPayload(protocol.Type(mt), messagePayload)

			channel <- packagedMessage
		}

	}()

	return channel
}
