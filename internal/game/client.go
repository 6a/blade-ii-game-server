// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package game implements the Blade II Online game server.
package game

import (
	"sync"
	"time"

	"github.com/6a/blade-ii-game-server/internal/connection"
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

// closeWaitPeriod is the time to wait between sending a close message, and closing a websocket.
const closeWaitPeriod = time.Second * 1

// GClient is a container for a websocket connection and its associated user data.
type GClient struct {

	// Database values for this client.
	DBID        uint64
	MatchID     uint64
	PublicID    string
	DisplayName string
	Avatar      uint8

	// Whether the server is currently expecting a move update from this client.
	WaitingForMove bool

	// A pointer to the websocket connection for this client.
	connection *connection.Connection

	// A pointer to the game server.
	server *Server

	// Whether this client is currently due to be disconnected.
	pendingKill bool

	// Mutex lock to protect the critical section that can occur when reading/writing to
	// pendingKill.
	killLock sync.Mutex
}

// StartEventLoop starts the send and receive pumps for the client, with a separate goroutine for each.
func (client *GClient) StartEventLoop() {
	go client.pollReceive()
	go client.pollSend()
}

// pollReceive loops forever, blocking until a new message from the websocket is available to read.
//
// On websocket error, the client will be added to the remove queue and the loop will break.
func (client *GClient) pollReceive() {
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
			client.server.Remove(client, protocol.WSCUnknownConnectionError, err.Error())
			break
		}
	}
}

// pollReceive loops forever, blocking until a new message from the websocket is ready to be sent.
//
// On websocket error, the client will be added to the remove queue and the loop will break.
func (client *GClient) pollSend() {
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
			client.server.Remove(client, protocol.WSCUnknownConnectionError, err.Error())
			break
		}
	}
}

// IsSameConnection returns true if the specified client is the same as this one.
func (client *GClient) IsSameConnection(other *GClient) bool {

	// Check to ensure that a null pointer is not checked - always returns false.
	if other == nil {
		return false
	}

	// Return true if the UUID's of the two clients are the same.
	return client.connection.UUID.Compare(other.connection.UUID) == 0
}

// SendMessage adds a message to the outbound queue.
func (client *GClient) SendMessage(message protocol.Message) {

	// Add this message to the outbound queue on the underlying websocket connection.
	client.connection.SendMessage(message)
}

// Close sends a message to the client, and closes the connection after a delay.
// The delay is asynchronous, as it is wrapped in a goroutine.
func (client *GClient) Close(message protocol.Message) {

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
func (client *GClient) isPendingKill() bool {

	// Lock the mutex lock, and then defer unlocking.
	client.killLock.Lock()
	defer client.killLock.Unlock()

	// Return the value of pendingKill. After the function exits, the lock will be
	// released.
	return client.pendingKill
}

// NewClient creates a and retruns a pointer to a new Client, and starts its
// message pumps in two seperate go routines.
func NewClient(wsconn *websocket.Conn, databaseID uint64, publicID string, displayname string, matchID uint64, avatar uint8, gameServer *Server) *GClient {
	connection := connection.NewConnection(wsconn)
	client := &GClient{
		DBID:           databaseID,
		PublicID:       publicID,
		DisplayName:    displayname,
		MatchID:        matchID,
		Avatar:         avatar,
		connection:     connection,
		server:         gameServer,
		WaitingForMove: false,
	}

	// Start the event loop for the new client.
	client.StartEventLoop()

	return client
}
