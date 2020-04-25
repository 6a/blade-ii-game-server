package game

import (
	"bytes"

	"github.com/6a/blade-ii-game-server/internal/protocol"
)

const clientDataDelimiter string = "."
const clientDataAndCardsDelimiter string = ":"

// Match is a wrapper for a matche's data and client connections etc
type Match struct {
	ID      uint64
	Client1 *GClient
	Client2 *GClient
	State   State
	Server  *Server
}

// Tick reads any incoming messages and passes outgoing messages to the queue
func (match *Match) Tick() {
	// Client 1
	for len(match.Client1.connection.ReceiveQueue) > 0 {
		message := match.Client1.connection.GetNextReceiveMessage()
		if message.Type == protocol.Type(protocol.WSMTText) {
			if message.Payload.Code == protocol.WSCMatchMoveUpdate {
				if match.isValidMove(message.Payload.Message) {
					match.Client2.SendMessage(protocol.NewMessageFromPayload(protocol.WSMTText, message.Payload))
				} else {
					match.Server.Remove(match.Client1, protocol.WSCMatchIllegalMove, "")

					// Record the result in the database
				}
			} else if message.Type == protocol.Type(protocol.WSCMatchForfeit) {
				match.Server.Remove(match.Client1, protocol.WSCMatchForfeit, "")

				// Record the result in the database
			} else if message.Type == protocol.Type(protocol.WSCMatchMessage) {

			}
		} else {
			// Handle non-text messages?
		}
	}
}

// TickClient performs the tick actions for the specified client
func (match *Match) TickClient(client *GClient) {

}

// IsFull returns true when the match is occupied by two players
func (match *Match) IsFull() bool {
	return match.Client1 != nil && match.Client2 != nil
}

// BroadCast sends the specified message to both clients
func (match *Match) BroadCast(message protocol.Message) {

	// Edge case - for cards being sent, we have to tag them so the client knows which of the decks is their own
	// Technically we need to convert them to hex strings but they will only ever be 0 and 1, so this is fine
	if message.Payload.Code == protocol.WSCMatchData {

		var client1Buffer bytes.Buffer
		var client2Buffer bytes.Buffer

		client1Buffer.WriteString(match.Client1.DisplayName)
		client1Buffer.WriteString(clientDataDelimiter)
		client1Buffer.WriteString(match.Client1.PublicID)
		client1Buffer.WriteString(clientDataAndCardsDelimiter)
		client1Buffer.WriteString("0")
		client1Buffer.WriteString(SerializedCardsDelimiter)
		client1Buffer.WriteString(message.Payload.Message)

		client1Buffer.WriteString(match.Client2.DisplayName)
		client1Buffer.WriteString(clientDataDelimiter)
		client1Buffer.WriteString(match.Client2.PublicID)
		client1Buffer.WriteString(clientDataAndCardsDelimiter)
		client1Buffer.WriteString("1")
		client1Buffer.WriteString(SerializedCardsDelimiter)
		client1Buffer.WriteString(message.Payload.Message)

		message.Payload.Message = client1Buffer.String()
		match.Client1.SendMessage(message)

		message.Payload.Message = client2Buffer.String()
		match.Client2.SendMessage(message)
	} else {
		match.Client1.SendMessage(message)
		match.Client2.SendMessage(message)
	}
}

func (match *Match) isValidMove(payloadMessage string) bool {

	return true
}
