package queue

import (
	"log"
	"time"

	"github.com/6a/blade-ii-game-server/internal/connection"
	"github.com/6a/blade-ii-game-server/internal/protocol"
	"github.com/gorilla/websocket"
)

// Client is a container for a websocket connection and its associate player data
type Client struct {
	ID              uint64
	UID             string
	MMR             int
	Ready           bool
	ReadyTime       time.Time
	IsReadyChecking bool
	connection      *connection.Connection
	queue           *Queue
}

// StartEventLoop is the event loop for this client (sends/receives messages)
func (client *Client) StartEventLoop() {
	go client.pollReceive()
	go client.pollSend()
}

func (client *Client) pollReceive() {
	for {
		err := client.connection.ReadMessage()
		if err != nil {
			log.Println("read error: ", err)
			client.queue.Remove(client)
			client.connection.Close()
			break
		}
	}
}

func (client *Client) pollSend() {
	for {
		message := client.connection.GetNextSendMessage()

		err := client.connection.WriteMessage(message)
		if err != nil {
			log.Println("write error: ", err)
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

		// For now we just relay all incoming messages
		client.connection.SendMessage(message)
	}

	// Update values
}

// SendMessage sends a message to the client
func (client *Client) SendMessage(message protocol.Message) {
	client.connection.SendMessage(message)
}

// NewClient creates a new Client
func NewClient(wsconn *websocket.Conn, uid string, mmr int, queue *Queue) Client {
	connection := connection.NewConnection(wsconn)

	return Client{
		connection: &connection,
		UID:        uid,
		MMR:        mmr,
		queue:      queue,
	}
}
