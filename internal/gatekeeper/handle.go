package gatekeeper

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

const timeout = time.Second * 5
const authDelimiter = ":"

// Handle waits for the new connection to send an auth protocol.
// Once received, it checks if the auth is valid, then retrieves the mmr of the
// specified account.
//
// If it does not receive an auth message within the timeout period, it drops the
// connection
func Handle(wsconn *websocket.Conn, mm *matchmaking.MatchMaking) {
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
		uid, b2code, err := checkAuth(res.Payload)
		if err != nil {
			connection.CloseConnection(wsconn, protocol.NewMessage(protocol.WSMTText, b2code, err.Error()))
			return
		}

		mmr, err := database.GetMMR(uid)
		if err != nil {
			connection.CloseConnection(wsconn, protocol.NewMessage(protocol.WSMTText, protocol.WSCUnknownError, err.Error()))
			return
		}

		mm.AddClient(wsconn, uid, mmr)
	case <-time.After(timeout):
		connection.CloseConnection(wsconn, protocol.NewMessage(protocol.WSCAuthNotReceived, protocol.WSCUnknownError, "Auth message not received"))
		return
	}
}

func checkAuth(payload protocol.Payload) (uid string, wscode protocol.B2Code, err error) {
	if payload.Code != protocol.WSCAuthRequest {
		return "", protocol.WSCAuthExpected, errors.New("Auth expected but received something else")
	}

	auth := strings.Split(payload.Message, authDelimiter)
	if len(auth) != 2 {
		return "", protocol.WSCAuthBadFormat, errors.New("Auth bad format")
	}

	uid, key := auth[0], auth[1]
	err = database.ValidateAuth(uid, key)
	if err != nil {
		// TODO filter by result (banned etc)
		return "", protocol.WSCAuthBadCredentials, errors.New("Credentials no invalid")
	}

	return uid, 0, nil
}
