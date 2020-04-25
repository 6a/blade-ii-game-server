package transactions

import (
	"time"

	"github.com/6a/blade-ii-game-server/internal/game"
	"github.com/6a/blade-ii-game-server/internal/protocol"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/matchmaking"
	"github.com/gorilla/websocket"
)

const connectionTimeOut = time.Second * 5

// HandleGSConnection waits for the new connection to send an auth protocol.
// Once received, it checks if the auth is valid, and then waits for the
// connection to send a match ID.
//
// If it does not receive an auth message and match ID within the timeout period, it drops the
// connection
func HandleGSConnection(wsconn *websocket.Conn, gs *game.Server) {
	inChannel := waitForMessageAsync(wsconn, 2)

	var id uint64
	var pid string
	var b2code protocol.B2Code
	var err error
	var authReceived bool = false

	for {
		select {
		case res := <-inChannel:
			if !authReceived {
				id, pid, b2code, err = checkAuth(res.Payload)
				if err != nil {
					Discard(wsconn, protocol.NewMessage(protocol.WSMTText, b2code, err.Error()))
					return
				}

				authReceived = true
			} else {
				matchID, b2code, err := validateMatch(id, res.Payload)
				if err != nil {
					Discard(wsconn, protocol.NewMessage(protocol.WSMTText, b2code, err.Error()))
					return
				}

				gs.AddClient(wsconn, id, pid, matchID)
				return
			}
		case <-time.After(connectionTimeOut):
			if !authReceived {
				Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCAuthNotReceived, "Auth not received"))
			} else {
				Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCMatchIDNotReceived, "Match ID not received"))
			}

			return
		}
	}
}

// HandleMMConnection waits for the new connection to send an auth protocol.
// Once received, it checks if the auth is valid, then retrieves the mmr of the
// specified account.
//
// If it does not receive an auth message within the timeout period, it drops the
// connection
func HandleMMConnection(wsconn *websocket.Conn, mm *matchmaking.Server) {
	authChannel := waitForMessageAsync(wsconn, 1)

	select {
	case res := <-authChannel:
		id, pid, b2code, err := checkAuth(res.Payload)
		if err != nil {
			Discard(wsconn, protocol.NewMessage(protocol.WSMTText, b2code, err.Error()))
			return
		}

		mmr, err := database.GetMMR(id)
		if err != nil {
			Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownError, err.Error()))
			return
		}

		mm.AddClient(wsconn, id, pid, mmr)
	case <-time.After(connectionTimeOut):
		Discard(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownError, "Auth message not received"))
		return
	}
}
