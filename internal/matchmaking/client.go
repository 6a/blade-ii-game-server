// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package matchmaking implements the Blade II Online matchmaking server.
package matchmaking

import (
	"sync"
	"time"

	"github.com/6a/blade-ii-game-server/internal/connection"
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

// closeWaitPeriod is the time to wait between sending a close message, and closing a websocket.
const closeWaitPeriod = time.Second * 1

// MMClient is a container for a websocket connection and its associate player data.
type MMClient struct {

	// Database values for this client.
	DBID     uint64
	PublicID string
	MMR      int

	// The clients own index within the matchmaking queue.
	QueueIndex uint64

	// Whether the client is ready (for ready checking).
	Ready bool

	// The time at which the client ready confirmation was received (for ready checking).
	ReadyTime time.Time

	// Whether the client is currently waiting for a ready confirmation (for ready checking).
	IsReadyChecking bool

	// Whether the other client is ready (for ready checking).
	AcceptMessageSentToOpponent bool

	// A pointer to the websocket connection for this client.
	connection *connection.Connection

	// A unique ID used for sorting - Should be set once connected
	ClientID uint64

	// A pointer to the matchmaking queue.
	queue *Queue

	// Whether this client is currently due to be disconnected.
	pendingKill bool

	// Mutex lock to protect the critical section that can occur when reading/writing to
	// pendingKill.
	killLock sync.Mutex
}

// StartEventLoop starts the send and receive pumps for the client, with a separate goroutine for each.
func (client *MMClient) StartEventLoop() {
	go client.pollReceive()
	go client.pollSend()
}

// pollReceive loops forever, blocking until a new message from the websocket is available to read.
//
// On websocket error, the client will be added to the remove queue and the loop will break.
func (client *MMClient) pollReceive() {
	for {

		// Block until a new message is received.
		err := client.connection.ReadMessage()

		// If the client is pending kill (most likely due to being terminated by another thread)
		// break out of the loop without doing anything.
		if client.isPendingKill() {
			break
		}

		// If the read function returned an error, remove this client from the server and
		// break out of the loop.
		if err != nil {
			client.queue.Remove(client, protocol.WSCUnknownConnectionError, err.Error())
			break
		}
	}
}

// pollReceive loops forever, blocking until a new message from the websocket is ready to be sent.
//
// On websocket error, the client will be added to the remove queue and the loop will break.
func (client *MMClient) pollSend() {
	for {

		// Block until a new outbound message is received.
		message := client.connection.GetNextOutboundMessage()

		// Attempt to write the message to the websocket.
		err := client.connection.WriteMessage(message)

		// If the client is pending kill (most likely due to being terminated by another thread)
		// break out of the loop without doing anything.
		if client.isPendingKill() {
			break
		}

		// If the write function returned an error, remove this client from the server and
		// break out of the loop.
		if err != nil {
			client.queue.Remove(client, protocol.WSCUnknownConnectionError, err.Error())
			break
		}

	}
}

// Tick processes all the work for this client.
func (client *MMClient) Tick() {

	// If the inbound message queue contains some data, read from it until it is empty.
	for len(client.connection.InboundMessageQueue) > 0 {

		// Read the next message from the inbound message queue.
		message := client.connection.GetNextInboundMessage()

		// If the message was a match making accept message, update the clients internal state
		// accordingly.
		if message.Payload.Code == protocol.WSCMatchMakingAccept {
			client.Ready = true
			client.ReadyTime = time.Now()
		}
	}
}

// SendMessage adds a message to the outbound queue.
func (client *MMClient) SendMessage(message protocol.Message) {

	// Add this message to the outbound queue on the underlying websocket connection.
	client.connection.SendMessage(message)
}

// Close sends a message to the client, and closes the connection after a delay.
// The delay is asynchronous, as it is wrapped in a goroutine.
func (client *MMClient) Close(message protocol.Message) {

	// Send the specified message to the client.
	client.SendMessage(message)

	// Using the client kill lock mutex to avoid race conditions, set pendingKill
	// to true, so that the next read/writes cause their respective pumps to exit.
	client.killLock.Lock()
	client.pendingKill = true
	client.killLock.Unlock()

	// Spin up a goroutine, which sleeps for a set amount for a set amount of time before closing
	// the websocket connection.
	go func() {
		time.Sleep(closeWaitPeriod)
		client.connection.WS.Close()
	}()
}

// isPendingKill is a helper function that returns true if this client is due to be killed.
//
// Uses a mutex lock to protect the critical section.
func (client *MMClient) isPendingKill() bool {

	// Lock the mutex lock, and then defer unlocking.
	client.killLock.Lock()
	defer client.killLock.Unlock()

	// Return the value of pendingKill. After the function exits, the lock will be
	// released.
	return client.pendingKill
}

// NewClient creates a and retruns a pointer to a new Client, and starts its
// message pumps in two seperate go routines.
func NewClient(wsconn *websocket.Conn, dbid uint64, pid string, mmr int, queue *Queue) *MMClient {
	connection := connection.NewConnection(wsconn)
	client := &MMClient{
		connection: connection,
		DBID:       dbid,
		PublicID:   pid,
		MMR:        mmr,
		queue:      queue,
	}

	// Start the event loop for the new client.
	client.StartEventLoop()

	return client
}
