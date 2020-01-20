package net

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Client is a container for a websocket connection and its associate player data
type Client struct {
	connection *Connection
	UID        string
	MMR        int
}

func (client *Client) readMessage() (WSMessageType, []byte, error) {
	mt, payload, err := client.connection.WS.ReadMessage()
	return WSMessageType(mt), payload, err
}

func (client *Client) writeMessage(messageType WSMessageType, message []byte) error {
	return client.connection.WS.WriteMessage(int(messageType), message)
}

// StartEventLoop is the event loop for this client (sends/receives messages)
func (client *Client) StartEventLoop() {
	go client.pollReceive()
	go client.pollSend()
}

func (client *Client) pollReceive() {
	for {
		mt, message, err := client.readMessage()
		if err != nil {
			log.Println("read error: ", err)
		}
		client.connection.AddToInQueue(MakeRawMessage(mt, message, err))
	}
}

func (client *Client) pollSend() {
	for {
		for len(client.connection.Out) > 0 {
			rawMessage := client.connection.PopOutQueue()

			err := client.writeMessage(rawMessage.Type, rawMessage.Payload)
			if err != nil {
				log.Println("write error: ", err)
			}

			log.Println(fmt.Sprintf("Client [%v] sent message:\n%s", client.UID, rawMessage))
		}
	}
}

// Tick reads any incoming messages and passes outgoing messages to the queue
func (client *Client) Tick() {
	for len(client.connection.In) > 0 {
		rawMessage := client.connection.PopInQueue()
		log.Println(fmt.Sprintf("Client [%v] received message:\n%s", client.UID, rawMessage))

		client.connection.AddToOutQueue(rawMessage)
	}
}

// NewClient creates a new Client
func NewClient(wsconn *websocket.Conn, uid string, mmr int) Client {

	connection := Connection{
		WS:     wsconn,
		Joined: time.Now(),
	}

	return Client{
		connection: &connection,
		UID:        uid,
		MMR:        mmr,
	}
}

// ClientPair is a light wrapper for a pair of client connections
type ClientPair struct {
	C1 *Client
	C2 *Client
}

// NewPair creates a new ClientPair
func NewPair(c1 *Client, c2 *Client) ClientPair {
	return ClientPair{
		C1: c1,
		C2: c2,
	}
}
