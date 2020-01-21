package net

// QueueCommandType is a typedef for commands, used to identify the purpose of a websocket message
type QueueCommandType uint16

// Queue Command types
const (
	QCTBroadcastMessage QueueCommandType = iota
	QCTDropAll
	QCTChangePollTime
)

// QueueCommand is a wrapper for a queue command and any accompanying data
type QueueCommand struct {
	Type QueueCommandType
	Data string
}
