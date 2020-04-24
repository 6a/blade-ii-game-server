package transactions

import (
	"errors"
	"strings"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

const authDelimiter = ":"

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
