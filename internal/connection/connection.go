package connection

import (
	"time"

	"github.com/rs/xid"

	"github.com/6a/blade-ii-game-server/internal/protocol"

	"github.com/gorilla/websocket"
)

const (
	// MessageBufferSize is the size of each clients message buffer (both directions)
	MessageBufferSize = 32

	// Maximum tduration to wait before a write is considered to have failed
	maximumWriteWait = time.Second * 8

	// Maximum duration to wait before a pong is considered to be MIA
	pongWait = maximumWriteWait * 2

	// Send a ping every (n) seconds
	pingPeriod = (pongWait * 8) / 10
)

// Connection is a wrapper for a websocket connection
type Connection struct {
	WS           *websocket.Conn
	Joined       time.Time
	Latency      time.Duration
	ReceiveQueue chan protocol.Message
	SendQueue    chan protocol.Message
	UUID         xid.ID
	pingTimer    *time.Timer
	lastPingTime time.Time
}

func (connection *Connection) init() {
	connection.ReceiveQueue = make(chan protocol.Message, MessageBufferSize)
	connection.SendQueue = make(chan protocol.Message, MessageBufferSize)

	// Set up pong handler
	connection.WS.SetReadDeadline(time.Now().Add(pongWait))
	connection.WS.SetPongHandler(connection.pongHandler)

	// Ticker
	connection.pingTimer = time.NewTimer(pingPeriod)

	// UUID
	connection.UUID = xid.New()
}

func (connection *Connection) pongHandler(pong string) error {
	connection.WS.SetReadDeadline(time.Now().Add(pongWait))

	connection.Latency = time.Now().Sub(connection.lastPingTime)

	connection.pingTimer.Reset(pingPeriod)

	return nil
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
	connection.SendQueue <- message
}

// GetNextReceiveMessage gets the next message from the client from the inbound message queue
func (connection *Connection) GetNextReceiveMessage() protocol.Message {
	return <-connection.ReceiveQueue
}

// GetNextSendMessage gets the next message from the outbound message queue
func (connection *Connection) GetNextSendMessage() protocol.Message {
	for {
		select {
		case message := <-connection.SendQueue:
			return message
		case <-connection.pingTimer.C:
			// Dont bother checking for errors as they will be picked up by the message pumps
			connection.lastPingTime = time.Now()
			_ = connection.WS.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(maximumWriteWait))
		}
	}
}

// Close closes the connection
func (connection *Connection) Close() error {
	connection.pingTimer.Stop()
	return connection.WS.Close()
}

// NewConnection creates a new connection
func NewConnection(wsconn *websocket.Conn) *Connection {
	connection := Connection{
		WS:      wsconn,
		Joined:  time.Now(),
		Latency: time.Second * 0,
	}

	connection.init()
	return &connection
}
