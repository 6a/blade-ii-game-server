// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package matchmaking implements the Blade II Online matchmaking server.
package matchmaking

import (
	"github.com/gorilla/websocket"
)

// Server is the matchmaking server itself
type Server struct {
	queue Queue
}

// AddClient takes a new client and their various data, wraps them up and adds them to the matchmaking server to be processed later.
func (ms *Server) AddClient(wsconn *websocket.Conn, dbid uint64, pid string, mmr int) {

	// Create a new client
	client := NewClient(wsconn, dbid, pid, mmr, &ms.queue)

	// Add it to the server.
	ms.queue.AddClient(client)
}

// Init initializes the matchmaking server including starting the internal loop.
func (ms *Server) Init() {

	// Start the queue (which is essentially the workhorse for the matchmaking server).
	ms.queue.Init()
}

// NewServer creates and returns a pointer to a new matchmaking server.
func NewServer() *Server {

	// Create a new matchmaking server.
	mms := Server{}

	// Initialize the matchmaking server.
	mms.Init()

	// Return a pointer to the newly created matchmaking server.
	return &mms
}
