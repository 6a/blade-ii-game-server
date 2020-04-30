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
	DisplayName string
	Avatar      uint8
	connection  *connection.Connection
	server      *Server
	pendingKill bool
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
			client.server.Remove(client, protocol.WSCUnknownConnectionError, err.Error())
			break
		}
	}
}

func (client *GClient) pollSend() {
	for {
		message := client.connection.GetNextSendMessage()

		err := client.connection.WriteMessage(message)

		if client.isPendingKill() {
			break
		}

		if err != nil {
			client.server.Remove(client, protocol.WSCUnknownConnectionError, err.Error())
			break
		}
	}
}

// IsSameConnection returns true if the passed in client is the same as the object that this function is called from
func (client *GClient) IsSameConnection(other *GClient) bool {
	return other != nil && client.connection.UUID.Compare(other.connection.UUID) == 0
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
	client.SendMessage(message)

	client.killLock.Lock()
	client.pendingKill = true
	client.killLock.Unlock()

	time.Sleep(closeWaitPeriod)
	client.connection.WS.Close()
}

func (client *GClient) isPendingKill() bool {
	client.killLock.Lock()
	defer client.killLock.Unlock()
	return client.pendingKill
}

// NewClient creates a new Client
func NewClient(wsconn *websocket.Conn, dbid uint64, pid string, displayname string, matchID uint64, avatar uint8, gameServer *Server) GClient {
	connection := connection.NewConnection(wsconn)

	return GClient{
		DBID:        dbid,
		PublicID:    pid,
		DisplayName: displayname,
		MatchID:     matchID,
		Avatar:      avatar,
		connection:  &connection,
		server:      gameServer,
	}
}
