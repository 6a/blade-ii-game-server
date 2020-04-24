package matchmaking

import (
	"github.com/gorilla/websocket"
)

// Server is the matchmaking server itself
type Server struct {
	queue Queue
}

// AddClient takes a new client and their various data, wraps them up and adds them to the matchmaking server
func (ms *Server) AddClient(wsconn *websocket.Conn, dbid uint64, pid string, mmr int) {
	client := NewClient(wsconn, dbid, pid, mmr, &ms.queue)
	client.StartEventLoop()
	ms.queue.Add(&client)
}

// Init initializes the matchmaking server including starting the internal loop
func (ms *Server) Init() {
	ms.queue.Start()
}

// NewServer creates a new matchmaking server
func NewServer() Server {
	ms := Server{}
	return ms
}
