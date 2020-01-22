package queue

import (
	"log"

	"github.com/6a/blade-ii-game-server/internal/connection"
	"github.com/gorilla/websocket"
)

// Client is a container for a websocket connection and its associate player data
type Client struct {
	connection *connection.Connection
	ID         uint64
	UID        string
	MMR        int
	queue      *Queue
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
		rawMessage := client.connection.GetNextReceiveMessage()

		// For now we just relay all incoming messages
		client.connection.SendMessage(connection.NewMessage(rawMessage.Type, connection.WSCInfo, string(rawMessage.Payload)))
	}

	// Update values
}

// SendMessage sends a message to the client
func (client *Client) SendMessage(message connection.Message) {
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
