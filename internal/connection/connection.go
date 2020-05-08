// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package connection implements a websocket connection wrapper with various helper functions.
package connection

import (
	"time"

	"github.com/rs/xid"

	"github.com/6a/blade-ii-game-server/internal/protocol"

	"github.com/gorilla/websocket"
)

const (
	// MessageBufferSize is the size of each clients message buffer (both directions).
	MessageBufferSize = 32

	// maximumWriteWait is the maximum duration to wait before a write is considered to have failed.
	maximumWriteWait = time.Second * 8

	// pongWait is the maximum duration to wait before a connection is considered to be dead due to no inbound traffic (pong or client message).
	pongWait = maximumWriteWait * 2

	// pingPeriod is the duration to wait after a ping is received, before sending another one.
	pingPeriod = (pongWait * 8) / 10
)

// Connection is a wrapper for a websocket connection.
type Connection struct {
	WS                   *websocket.Conn       // The websocket connection itself.
	Joined               time.Time             // The time at which the connection was created.
	Latency              time.Duration         // The current latency of the connection.
	InboundMessageQueue  chan protocol.Message // Inbound message queue - received messages are parked here until removed by a read pump.
	OutboundMessageQueue chan protocol.Message // Outbound message queue - messages to be sent are parked here until removed by a write pump.
	UUID                 xid.ID                // A unique ID for this connection.
	pingTimer            *time.Timer           // A timer use to handle the ping pong keep-alive.
	lastPingTime         time.Time             // The time at which the most recent ping was sent.
}

// init initialises a connection object, setting up the internal ping/pong handler, message queues, and assigning a UUID.
func (connection *Connection) init() {

	// Initialise the send and receive queues.
	connection.InboundMessageQueue = make(chan protocol.Message, MessageBufferSize)
	connection.OutboundMessageQueue = make(chan protocol.Message, MessageBufferSize)

	// Set up pong handler.
	connection.WS.SetReadDeadline(time.Now().Add(pongWait))
	connection.WS.SetPongHandler(connection.pongHandler)

	// Set up the ticker that dictates when pings should be sent.
	connection.pingTimer = time.NewTimer(pingPeriod)

	// Generate and assign a UUID.
	connection.UUID = xid.New()
}

// pongHandler handles pong messages from the client.
func (connection *Connection) pongHandler(pong string) error {

	// Reset the read deadline based on the current time.
	connection.WS.SetReadDeadline(time.Now().Add(pongWait))

	// Calculate the latency of the connection (round trip).
	connection.Latency = time.Now().Sub(connection.lastPingTime)

	// Reset the ping timer, so that it will fire again later.
	connection.pingTimer.Reset(pingPeriod)

	// Never return an error - Timeouts are handled by the standard read deadline handler.
	return nil
}

// ReadMessage synchronously retreives messages from the websocket.
func (connection *Connection) ReadMessage() error {

	// Wait until the websocket read function returns, and inspect the return values.
	mt, payload, err := connection.WS.ReadMessage()
	if err != nil {
		return err
	}

	// If the message was read successfully, convert it into the internal message container for use.
	// within the application.
	messagePayload := protocol.NewPayloadFromBytes(payload)
	packagedMessage := protocol.NewMessageFromPayload(protocol.Type(mt), messagePayload)

	// Add the packaged message data to the receive queue, ready to be read by the application.
	connection.InboundMessageQueue <- packagedMessage

	return nil
}

// WriteMessage synchronously sends messages down the websocket.
func (connection *Connection) WriteMessage(message protocol.Message) error {

	// Write a message to the websocket based on the passed in message.
	return connection.WS.WriteMessage(int(message.Type), message.GetPayloadBytes())
}

// SendMessage asynchronously sends a message down the websocket.
func (connection *Connection) SendMessage(message protocol.Message) {

	// Add the message to the outbound queue.
	connection.OutboundMessageQueue <- message
}

// GetNextInboundMessage gets the next message from the inbound message queue.
// Blocks when the queue is empty, so check the queue's length if you don't want to wait.
func (connection *Connection) GetNextInboundMessage() protocol.Message {

	// Deqeue and return the message.
	return <-connection.InboundMessageQueue
}

// GetNextOutboundMessage gets the next message from the outbound message queue.
// Blocks when the queue is empty, so check the queue's length if you don't want to wait.
func (connection *Connection) GetNextOutboundMessage() protocol.Message {

	// Wait for a message to be added to the outbound message queue.
	// A loop + select is used so that the ping timer can interrupt the queue read if its blocking,
	// allowing for a ping message to be sent. Once sent, control returns to the queue read, until
	// either a message is added to the outbound queue, or the ping timer interrupts the read again,
	// ad infinitum...
	for {
		select {

		// Blocks until read.
		case message := <-connection.OutboundMessageQueue:
			return message

		// The ping timer is able to bypass the blocked queue read, enabling the ping message to be sent.
		case <-connection.pingTimer.C:

			// Store the time at which this ping was sent (for calculating latency).
			connection.lastPingTime = time.Now()

			// Write a ping message. Dont bother checking for errors as they will be detected when the websocket is
			// next written to / read from.
			connection.WS.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(maximumWriteWait))
		}
	}
}

// Close closes the connection.
func (connection *Connection) Close() error {

	// Stop the ping timer to avoid it triggering while the websocket is in an invalid state due to being closed,
	// or being in the process of closing etc..
	connection.pingTimer.Stop()

	// Close the websocket connection, and return any errors.
	return connection.WS.Close()
}

// NewConnection creates a new connection.
func NewConnection(wsconn *websocket.Conn) *Connection {

	// Create a new connection, with the provided websocket connection.
	connection := Connection{
		WS:      wsconn,
		Joined:  time.Now(),
		Latency: time.Second * 0,
	}

	// Initialise, and then return the connection.
	connection.init()
	return &connection
}
