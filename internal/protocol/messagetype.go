package protocol

// Type is a type definition for websocket message types
type Type uint16

// Types of Websocket Message
const (
	WSMTContinuation Type = 0
	WSMTText              = 1
	WSMTBinary            = 2
)
