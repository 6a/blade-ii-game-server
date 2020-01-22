package gatekeeper

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/6a/blade-ii-game-server/internal/connection"
	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/matchmaking"
	"github.com/gorilla/websocket"
)

const timeout = time.Second * 5
const authDelimiter = ":"

// Handle waits for the new connection to send an auth message.
// Once received, it checks if the auth is valid, then retrieves the mmr of the
// specified account.
//
// If it does not receive an auth message within the timeout period, it drops the
// connection
func Handle(wsconn *websocket.Conn, mm *matchmaking.MatchMaking) {
	readyChannel := make(chan connection.RawMessage, 1)
	go func() {
		mt, payload, err := wsconn.ReadMessage()
		if err != nil {
			connection.CloseConnection(wsconn, connection.WSCUnknownError, err.Error())
			return
		}

		msg := connection.NewRawMessage(connection.WSMessageType(mt), payload)
		readyChannel <- msg
	}()

	select {
	case res := <-readyChannel:
		uid, wscode, err := checkAuth(res.Payload)
		if err != nil {
			connection.CloseConnection(wsconn, wscode, err.Error())
			return
		}

		mmr, err := database.GetMMR(uid)
		if err != nil {
			connection.CloseConnection(wsconn, connection.WSCUnknownError, err.Error())
			return
		}

		mm.AddClient(wsconn, uid, mmr)
	case <-time.After(timeout):
		connection.CloseConnection(wsconn, connection.WSCAuthNotReceived, "Auth message not received")
		return
	}
}

func checkAuth(payloadBytes []byte) (uid string, wscode connection.WSCode, err error) {
	payload := connection.Payload{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return "", connection.WSCAuthBadFormat, errors.New("Auth bad format")
	}

	if payload.Code != connection.WSCAuthRequest {
		return "", connection.WSCAuthExpected, errors.New("Auth expected but received something else")
	}

	auth := strings.Split(payload.Message, authDelimiter)
	if len(auth) != 2 {
		return "", connection.WSCAuthBadFormat, errors.New("Auth bad format")
	}

	uid, key := auth[0], auth[1]
	err = database.ValidateAuth(uid, key)
	if err != nil {
		// TODO filter by result (banned etc)
		return "", connection.WSCAuthBadCredentials, errors.New("Credentials no invalid")
	}

	return uid, 0, nil
}
