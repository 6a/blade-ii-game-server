package net

import (
	"github.com/gorilla/websocket"
)

// MatchMaking is the matchmaking queue itself
type MatchMaking struct {
	queue Queue
}

// Start initialises the matchmaking queue
func (mm *MatchMaking) Start() {
	mm.queue.Start()
}

// AddClient takes a new client and their various data, wraps them up and adds them to the matchmaking queue
func (mm *MatchMaking) AddClient(wsconn *websocket.Conn, uid string, mmr int) {
	client := NewClient(wsconn, uid, mmr, &mm.queue)
	client.StartEventLoop()

	client.SendMessage(NewMessage(WSMTText, WSCAuthSuccess, "Added to matchmaking queue"))

	mm.queue.Add(&client)
}

// NewMatchMaking creates a new matchmaking queue
func NewMatchMaking() MatchMaking {
	return MatchMaking{}
}
