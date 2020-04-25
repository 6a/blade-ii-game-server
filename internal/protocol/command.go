package protocol

// Queue Command types
const (
	QCTBroadcastMessage uint16 = iota
	QCTDropAll
	QCTChangePollTime
)

// Command is a wrapper for a queue command and any accompanying data
type Command struct {
	Type uint16
	Data string
}
