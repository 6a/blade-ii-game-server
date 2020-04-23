package queue

import (
	"sync"
	"time"

	"github.com/6a/blade-ii-game-server/internal/connection"
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

const closeWaitPeriod = time.Second * 1

// Client is a container for a websocket connection and its associate player data
type Client struct {
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
}

// StartEventLoop is the event loop for this client (sends/receives messages)
func (client *Client) StartEventLoop() {
	go client.pollReceive()
	go client.pollSend()
}

func (client *Client) pollReceive() {
	for {
		err := client.connection.ReadMessage()

		if client.isPendingKill() {
			break
		}

		if err != nil {
			client.queue.Remove(client, protocol.WSCUnknownError, err.Error())
			break
		}
	}
}

func (client *Client) pollSend() {
	for {
		message := client.connection.GetNextSendMessage()

		if client.isPendingKill() {
			break
		}

		err := client.connection.WriteMessage(message)
		if err != nil {
			client.queue.Remove(client, protocol.WSCUnknownError, err.Error())
			break
		}
	}
}

// Tick reads any incoming messages and passes outgoing messages to the queue
func (client *Client) Tick() {
	// Process receive queue
	for len(client.connection.ReceiveQueue) > 0 {
		message := client.connection.GetNextReceiveMessage()

		if message.Payload.Code == protocol.WSCMatchMakingReady {
			client.Ready = true
			client.ReadyTime = time.Now()
		}
	}
}

// SendMessage sends a message to the client
func (client *Client) SendMessage(message protocol.Message) {
	client.connection.SendMessage(message)
}

// Close closes a websocket connection immediately after sending the specified message
func (client *Client) Close(message protocol.Message) {
	client.killLock.Lock()
	client.PendingKill = true
	client.killLock.Unlock()

	client.SendMessage(message)

	time.Sleep(closeWaitPeriod)
	client.connection.WS.Close()
}

func (client *Client) isPendingKill() bool {
	client.killLock.Lock()
	defer client.killLock.Unlock()
	return client.PendingKill
}

// NewClient creates a new Client
func NewClient(wsconn *websocket.Conn, dbid uint64, pid string, mmr int, queue *Queue) Client {
	connection := connection.NewConnection(wsconn)

	return Client{
		connection: &connection,
		DBID:       dbid,
		PublicID:   pid,
		MMR:        mmr,
		queue:      queue,
	}
}
