// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package transactions implements handlers for various interactions with raw websocket connections,
// before they are packaged and added to the server.
package transactions

import (
	"errors"
	"strconv"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

// validateMatch checks if the match details contained in the payload, represent a match that is valid, and that
// the user with the specified database ID is a participant in the match. Returns an error if invalid, or if
// there was a database error.
func validateMatch(databaseID uint64, payload protocol.Payload) (matchID uint64, wscode protocol.B2Code, err error) {

	// Return an error immediately if the payload code was not the correct type.
	if payload.Code != protocol.WSCMatchID {
		return matchID, protocol.WSCMatchIDExpected, errors.New("Match ID expected but received something else")
	}

	// Attempt to parse the payload message into a uint64. Return an error if the parsing failed.
	matchID, err = strconv.ParseUint(payload.Message, 10, 64)
	if err != nil {
		return matchID, protocol.WSCMatchIDBadFormat, errors.New("Match ID format invalid or missing")
	}

	// Expiry check here
	// TODO impl expiry check

	// Check if the specified match exists, and the user with the specified database ID is part of it.
	// An error being returned indicates that the query failed or there was a database error. If valid
	// is false, then the match details were invalid.
	valid, err := database.ValidateMatch(databaseID, matchID)
	if err != nil {
		return matchID, protocol.WSCMatchInvalid, err
	} else if !valid {
		return matchID, protocol.WSCMatchInvalid, errors.New("Could not find a valid match with the specified details")
	}

	// Reaching this point means the match is valid.
	return matchID, wscode, err
}
