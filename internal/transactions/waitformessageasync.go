package transactions

import (
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

func waitForMessageAsync(wsconn *websocket.Conn) chan protocol.Message {
	channel := make(chan protocol.Message, 1)

	go func() {
		mt, payload, err := wsconn.ReadMessage()
		if err != nil {
			Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownError, err.Error()))
			return
		}

		messagePayload := protocol.NewPayloadFromBytes(payload)
		packagedMessage := protocol.NewMessageFromPayload(protocol.Type(mt), messagePayload)

		channel <- packagedMessage
	}()

	return channel
}
