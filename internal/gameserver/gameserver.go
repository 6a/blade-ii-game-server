package gameserver

import (
	"github.com/6a/blade-ii-game-server/internal/queue"
	"github.com/gorilla/websocket"
)

// Server is the game server itself
type Server struct {
	queue queue.MatchMakingQueue
}

// AddClient takes a new client and their various data, wraps them up and adds them to the game server
func (gs *Server) AddClient(wsconn *websocket.Conn, dbid uint64, pid string, matchID uint64) {
	// client := queue.NewClient(wsconn, dbid, pid, mmr, &mm.queue)
	// client.StartEventLoop()
	// mm.queue.Add(&client)
}

// Init initializes the game server including starting the internal loop
func (gs *Server) Init() {
	gs.queue.Start()
}

// NewGameServer creates a new game server
func NewGameServer() Server {
	gs := Server{}
	return gs
}
