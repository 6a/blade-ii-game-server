package net

import (
	"time"

	"github.com/gorilla/websocket"
)

// Connection is a wrapper for a websocket connection
type Connection struct {
	WS      *websocket.Conn
	Joined  time.Time
	Latency uint16
	In      []RawMessage
	Out     []RawMessage
}

func (c *Connection) AddToInQueue(msg RawMessage) {
	c.In = append(c.In, msg)
}

func (c *Connection) AddToOutQueue(msg RawMessage) {
	c.Out = append(c.Out, msg)
}

func (c *Connection) PopInQueue() RawMessage {
	rawMessage := c.In[0]
	c.In = c.In[1:]

	return rawMessage
}

func (c *Connection) PopOutQueue() RawMessage {
	rawMessage := c.Out[0]
	c.Out = c.Out[1:]

	return rawMessage
}
