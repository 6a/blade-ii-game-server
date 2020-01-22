package matchmaking

import (
	"github.com/6a/blade-ii-game-server/internal/queue"
	"github.com/gorilla/websocket"
)

// MatchMaking is the matchmaking queue itself
type MatchMaking struct {
	queue queue.Queue
}

// AddClient takes a new client and their various data, wraps them up and adds them to the matchmaking queue
func (mm *MatchMaking) AddClient(wsconn *websocket.Conn, uid string, mmr int) {
	client := queue.NewClient(wsconn, uid, mmr, &mm.queue)
	client.StartEventLoop()

	mm.queue.Add(&client)
}

// Init initializes the matchmaking queue including starting the internal loop
// Returns instantly after wrapping the logic in a goroutine
func (mm *MatchMaking) Init() {
	go mm.queue.Start()
}

// NewMatchMaking creates a new matchmaking queue, and starts a goroutine that runs the main internal loop
func NewMatchMaking() MatchMaking {
	mm := MatchMaking{}
	return mm
}
