package gameserver

import (
	"github.com/6a/blade-ii-game-server/internal/queue"
	"github.com/gorilla/websocket"
)

// GameServer is the matchmaking queue itself
type GameServer struct {
	queue queue.MatchMakingQueue
}

// AddClient takes a new client and their various data, wraps them up and adds them to the matchmaking queue
func (mm *GameServer) AddClient(wsconn *websocket.Conn, dbid uint64, pid string, mmr int) {
	client := queue.NewClient(wsconn, dbid, pid, mmr, &mm.queue)
	client.StartEventLoop()
	mm.queue.Add(&client)
}

// Init initializes the matchmaking queue including starting the internal loop
// Returns instantly after wrapping the logic in a goroutine
func (mm *GameServer) Init() {
	mm.queue.Start()
}

// NewMatchMaking creates a new matchmaking queue, and starts a goroutine that runs the main internal loop
func NewMatchMaking() GameServer {
	mm := GameServer{}
	return mm
}
