package gatekeeper

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/net"
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
func Handle(wsconn *websocket.Conn, mm *net.MatchMaking) {
	readyChannel := make(chan net.RawMessage, 1)
	go func() {
		mt, payload, err := wsconn.ReadMessage()
		if err != nil {
			net.CloseConnection(wsconn, net.WSCUnknownError, err.Error())
			return
		}

		msg := net.NewRawMessage(net.WSMessageType(mt), payload)
		readyChannel <- msg
	}()

	select {
	case res := <-readyChannel:
		uid, wscode, err := checkAuth(res.Payload)
		if err != nil {
			net.CloseConnection(wsconn, wscode, err.Error())
			return
		}

		mmr, err := database.GetMMR(uid)
		if err != nil {
			net.CloseConnection(wsconn, net.WSCUnknownError, err.Error())
			return
		}

		mm.AddClient(wsconn, uid, mmr)
	case <-time.After(timeout):
		net.CloseConnection(wsconn, net.WSCAuthNotReceived, "Auth message not received")
		return
	}
}

func checkAuth(payloadBytes []byte) (uid string, wscode net.WSCode, err error) {
	payload := net.Payload{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return "", net.WSCAuthBadFormat, errors.New("Auth bad format")
	}

	if payload.Code != net.WSCAuthRequest {
		return "", net.WSCAuthExpected, errors.New("Auth expected but received something else")
	}

	auth := strings.Split(payload.Message, authDelimiter)
	if len(auth) != 2 {
		return "", net.WSCAuthBadFormat, errors.New("Auth bad format")
	}

	uid, key := auth[0], auth[1]
	err = database.ValidateAuth(uid, key)
	if err != nil {
		// TODO filter by result (banned etc)
		return "", net.WSCAuthBadCredentials, errors.New("Credentials no invalid")
	}

	return uid, 0, nil
}
