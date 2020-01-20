package mm

import (
	"time"

	"github.com/6a/blade-ii-game-server/internal/net"
)

const (
	pollTime = 100 * time.Millisecond
)

var queue Queue

func init() {
	go poll()
}

func poll() {
	for {
		queue.Tick()
		time.Sleep(pollTime)
	}
}

// JoinQueue takes the client object and adds it to the matchmaking queue
func JoinQueue(client *net.Client) {
	client.StartEventLoop()
	queue.Add(client)
}
