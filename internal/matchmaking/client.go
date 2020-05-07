package matchmaking

import (
	"sync"
	"time"

	"github.com/6a/blade-ii-game-server/internal/connection"
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

const closeWaitPeriod = time.Second * 1

// MMClient is a container for a websocket connection and its associate player data
type MMClient struct {
	DBID            uint64
	PublicID        string
	QueueID         uint64
	MMR             int
	Ready           bool
	ReadyTime       time.Time
	IsReadyChecking bool
	connection      *connection.Connection
	queue           *Queue
	PendingKill     bool
	killLock        sync.Mutex
	OpponentReady   bool
}

// StartEventLoop is the event loop for this client (sends/receives messages)
func (client *MMClient) StartEventLoop() {
	go client.pollReceive()
	go client.pollSend()
}

func (client *MMClient) pollReceive() {
	for {
		err := client.connection.ReadMessage()

		if err != nil {
			client.queue.Remove(client, protocol.WSCUnknownConnectionError, err.Error())
			break
		}

		if client.isPendingKill() {
			break
		}
	}
}

func (client *MMClient) pollSend() {
	for {
		message := client.connection.GetNextOutboundMessage()

		err := client.connection.WriteMessage(message)
		if err != nil {
			client.queue.Remove(client, protocol.WSCUnknownConnectionError, err.Error())
			break
		}

		if client.isPendingKill() {
			break
		}
	}
}

// Tick reads any incoming messages and passes outgoing messages to the queue
func (client *MMClient) Tick() {
	// Process receive queue
	for len(client.connection.InboundMessageQueue) > 0 {
		message := client.connection.GetNextReceiveMessage()

		if message.Payload.Code == protocol.WSCMatchMakingAccept {
			client.Ready = true
			client.ReadyTime = time.Now()
		}
	}
}

// SendMessage sends a message to the client
func (client *MMClient) SendMessage(message protocol.Message) {
	client.connection.SendMessage(message)
}

// Close closes a websocket connection immediately after sending the specified message
func (client *MMClient) Close(message protocol.Message) {
	client.killLock.Lock()
	client.PendingKill = true
	client.killLock.Unlock()

	client.SendMessage(message)

	go func() {
		time.Sleep(closeWaitPeriod)
		client.connection.WS.Close()
	}()
}

func (client *MMClient) isPendingKill() bool {
	client.killLock.Lock()
	defer client.killLock.Unlock()
	return client.PendingKill
}

// NewClient creates a new Client
func NewClient(wsconn *websocket.Conn, dbid uint64, pid string, mmr int, queue *Queue) MMClient {
	connection := connection.NewConnection(wsconn)

	return MMClient{
		connection: connection,
		DBID:       dbid,
		PublicID:   pid,
		MMR:        mmr,
		queue:      queue,
	}
}
