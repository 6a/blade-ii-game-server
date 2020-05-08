// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package transactions implements handlers for various interactions with raw websocket connections,
// before they are packaged and added to the server.
package transactions

import (
	"errors"
	"strings"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

const (

	// authDelimiter is the delimiter that is used to separate the public id and auth token in an auth message.
	authDelimiter = ":"

	// expectedAuthArraySize is the expected size of the array output of splitting the auth message payload.
	expectedAuthArraySize = 2
)

// checkAuth attempts to extract the credentials from a payload, returning the database and public ID for the
// user for which the credentials matched. If there was an error, an errorcode is returned as well as an error.
func checkAuth(payload protocol.Payload) (databaseID uint64, publicID string, b2ErrorCode protocol.B2Code, err error) {

	// If the payload code was not that of an auth request, return immedaitely with an error.
	if payload.Code != protocol.WSCAuthRequest {
		return databaseID, publicID, protocol.WSCAuthExpected, errors.New("Auth expected but received something else")
	}

	// Attempt to split the payload string into an array containing a public ID and an auth token.
	auth := strings.Split(payload.Message, authDelimiter)

	// If the output array is not the right size, return immedaitely with an error.
	if len(auth) != expectedAuthArraySize {
		return databaseID, publicID, protocol.WSCAuthBadFormat, errors.New("Auth bad format")
	}

	// Create some local variables for each auth component for clarity.
	publicID, authToken := auth[0], auth[1]

	// Attempt to validate the credentials.
	databaseID, err = database.ValidateAuth(publicID, authToken)

	// If there was a database error, return immedaitely with an error, as it means that either there
	// was a problem accessing the database, or the credentials were invalid, or the account was banned
	// etc..
	if err != nil {
		// TODO filter by result (banned etc)
		return databaseID, publicID, protocol.WSCAuthBadCredentials, err
	}

	// By reaching this point, auth should be confirmed as valid, so return the database ID and the public
	// ID, with no error code or error.
	return databaseID, publicID, 0, nil
}
