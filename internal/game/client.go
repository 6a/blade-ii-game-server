package game

import (
	"sync"
	"time"

	"github.com/6a/blade-ii-game-server/internal/connection"
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

const closeWaitPeriod = time.Second * 1

// GClient is a container for a websocket connection and its associated player data
type GClient struct {
	DBID        uint64
	MatchID     uint64
	PublicID    string
	connection  *connection.Connection
	server      *Server
	PendingKill bool
	killLock    sync.Mutex
}

// StartEventLoop is the event loop for this client (sends/receives messages)
func (client *GClient) StartEventLoop() {
	go client.pollReceive()
	go client.pollSend()
}

func (client *GClient) pollReceive() {
	for {
		err := client.connection.ReadMessage()

		if client.isPendingKill() {
			break
		}

		if err != nil {
			// client.server.Remove(client, protocol.WSCUnknownError, err.Error())
			break
		}
	}
}

func (client *GClient) pollSend() {
	for {
		message := client.connection.GetNextSendMessage()

		if client.isPendingKill() {
			break
		}

		err := client.connection.WriteMessage(message)
		if err != nil {
			// client.server.Remove(client, protocol.WSCUnknownError, err.Error())
			break
		}
	}
}

// Tick reads any incoming messages and passes outgoing messages to the queue
func (client *GClient) Tick() {
	// Process receive queue
	for len(client.connection.ReceiveQueue) > 0 {
		// message := client.connection.GetNextReceiveMessage()
	}
}

// SendMessage sends a message to the client
func (client *GClient) SendMessage(message protocol.Message) {
	client.connection.SendMessage(message)
}

// Close closes a websocket connection immediately after sending the specified message
func (client *GClient) Close(message protocol.Message) {
	client.killLock.Lock()
	client.PendingKill = true
	client.killLock.Unlock()

	client.SendMessage(message)

	time.Sleep(closeWaitPeriod)
	client.connection.WS.Close()
}

func (client *GClient) isPendingKill() bool {
	client.killLock.Lock()
	defer client.killLock.Unlock()
	return client.PendingKill
}

// NewClient creates a new Client
func NewClient(wsconn *websocket.Conn, dbid uint64, pid string, mmr int, gameServer *Server) GClient {
	connection := connection.NewConnection(wsconn)

	return GClient{
		connection: &connection,
		DBID:       dbid,
		PublicID:   pid,
		server:     gameServer,
	}
}
