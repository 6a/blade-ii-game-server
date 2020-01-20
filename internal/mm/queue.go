package mm

import (
	"github.com/6a/blade-ii-game-server/internal/net"
)

// Queue is a wrapper for the matchmaking queue
type Queue struct {
	queue []*net.Client
}

// Add a client to the queue
func (queue *Queue) Add(client *net.Client) {
	queue.queue = append(queue.queue, client)
}

// Tick updates every client in order
func (queue *Queue) Tick() {
	for index := 0; index < len(queue.queue); index++ {
		queue.queue[index].Tick()
	}
}

// MatchMake matches people in the queue based on various factors, such as mmr, latency, and waiting time
func (queue *Queue) MatchMake() {
	// Matchmaking algo goes here. For now just return all the pairs we can

	// var pairs = []net.ClientPair{}

}
