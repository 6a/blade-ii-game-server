package connection

import (
	"time"

	"github.com/6a/blade-ii-matchmaking-server/internal/protocol"

	"github.com/gorilla/websocket"
)

// MessageBufferSize is the size of each clients message buffer (both directions)
const MessageBufferSize = 32

// Connection is a wrapper for a websocket connection
type Connection struct {
	WS           *websocket.Conn
	Joined       time.Time
	Latency      uint16
	ReceiveQueue chan protocol.Message
	Out          chan protocol.Message
}

func (connection *Connection) init() {
	connection.ReceiveQueue = make(chan protocol.Message, MessageBufferSize)
	connection.Out = make(chan protocol.Message, MessageBufferSize)
}

// ReadMessage synchronously retreives messages from the websocket
func (connection *Connection) ReadMessage() error {
	mt, payload, err := connection.WS.ReadMessage()
	if err != nil {
		return err
	}

	messagePayload := protocol.NewPayloadFromBytes(payload)
	packagedMessage := protocol.NewMessageFromPayload(protocol.Type(mt), messagePayload)
	connection.ReceiveQueue <- packagedMessage

	return nil
}

// WriteMessage synchronously sends messages down websocket
func (connection *Connection) WriteMessage(message protocol.Message) error {
	return connection.WS.WriteMessage(int(message.Type), message.GetPayloadBytes())
}

// SendMessage sends a messages (in reality it adds it to a queue and it is sent shortly after)
func (connection *Connection) SendMessage(message protocol.Message) {
	connection.Out <- message
}

// GetNextReceiveMessage gets the next message from the client from the inbound message queue
func (connection *Connection) GetNextReceiveMessage() protocol.Message {
	return <-connection.ReceiveQueue
}

// GetNextSendMessage gets the next message from the outbound message queue
func (connection *Connection) GetNextSendMessage() protocol.Message {
	return <-connection.Out
}

// Close closes the connection
func (connection *Connection) Close() error {
	return connection.WS.Close()
}

// NewConnection creates a new connection
func NewConnection(wsconn *websocket.Conn) Connection {
	connection := Connection{
		WS:     wsconn,
		Joined: time.Now(),
	}

	connection.init()
	return connection
}
