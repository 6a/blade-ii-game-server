package transactions

import (
	"errors"
	"strconv"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

func validateMatch(clientID uint64, payload protocol.Payload) (matchID uint64, wscode protocol.B2Code, err error) {
	if payload.Code != protocol.WSCMatchID {
		return matchID, protocol.WSCMatchIDExpected, errors.New("Match ID expected but received something else")
	}

	matchID, err = strconv.ParseUint(payload.Message, 10, 64)
	if err != nil {
		return matchID, protocol.WSCMatchIDBadFormat, errors.New("Match ID format invalid or missing")
	}

	// Expiry check here
	// TODO impl expiry check

	valid, err := database.ValidateMatch(clientID, matchID)
	if err != nil {
		return matchID, protocol.WSCMatchInvalid, err
	} else if !valid {
		return matchID, protocol.WSCMatchInvalid, errors.New("Could not find a valid match with the specified details")
	}

	return matchID, wscode, err
}
