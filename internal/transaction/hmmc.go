package transaction

import (
	"errors"
	"strings"
	"time"

	"github.com/6a/blade-ii-game-server/internal/connection"
	"github.com/6a/blade-ii-game-server/internal/protocol"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/matchmaking"
	"github.com/gorilla/websocket"
)

const mmConnectTimeout = time.Second * 5
const authDelimiter = ":"

// HandleMMConnection waits for the new connection to send an auth protocol.
// Once received, it checks if the auth is valid, then retrieves the mmr of the
// specified account.
//
// If it does not receive an auth message within the timeout period, it drops the
// connection
func HandleMMConnection(wsconn *websocket.Conn, mm *matchmaking.MatchMaking) {
	readyChannel := make(chan protocol.Message, 1)
	go func() {
		mt, payload, err := wsconn.ReadMessage()
		if err != nil {
			connection.CloseConnection(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownError, err.Error()))
			return
		}

		messagePayload := protocol.NewPayloadFromBytes(payload)
		packagedMessage := protocol.NewMessageFromPayload(protocol.Type(mt), messagePayload)

		readyChannel <- packagedMessage
	}()

	select {
	case res := <-readyChannel:
		id, pid, b2code, err := checkAuth(res.Payload)
		if err != nil {
			connection.CloseConnection(wsconn, protocol.NewMessage(protocol.WSMTText, b2code, err.Error()))
			return
		}

		mmr, err := database.GetMMR(id)
		if err != nil {
			connection.CloseConnection(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownError, err.Error()))
			return
		}

		mm.AddClient(wsconn, pid, mmr)
	case <-time.After(mmConnectTimeout):
		connection.CloseConnection(wsconn, protocol.NewMessage(protocol.WSCAuthNotReceived, protocol.WSCUnknownError, "Auth message not received"))
		return
	}
}

func checkAuth(payload protocol.Payload) (id uint64, pid string, wscode protocol.B2Code, err error) {
	if payload.Code != protocol.WSCAuthRequest {
		return id, pid, protocol.WSCAuthExpected, errors.New("Auth expected but received something else")
	}

	auth := strings.Split(payload.Message, authDelimiter)
	if len(auth) != 2 {
		return id, pid, protocol.WSCAuthBadFormat, errors.New("Auth bad format")
	}

	pid, key := auth[0], auth[1]
	id, err = database.ValidateAuth(pid, key)
	if err != nil {
		// TODO filter by result (banned etc)
		return id, pid, protocol.WSCAuthBadCredentials, err
	}

	return id, pid, 0, nil
}
