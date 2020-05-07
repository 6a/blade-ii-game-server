// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package database provides an interface through which the application can interact with a database.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql driver - Isn't explicitly used, so imported with no label.
)

// authExpiryGracePeriod defines the minimum duration of validity remaining for a auth token before it's considered invalid.
// This exists so that we can avoid race conditions that could occur if a token is changed in the database during an auth
// check.
const authExpiryGracePeriod = time.Minute * 10

var (
	// db is a pointer to this packages single instance of a database connection.
	db *sql.DB

	// envvars is a container for all of the environment variables used by the database package.
	envvars EnvironmentVariables

	// pstatements is a container for all of the prepared staments used by the database package.
	pstatements PreparedStatements
)

// Init should be called at the start of the function. It opens a connection to the database
// based on the parameters defined by environment variables, as specified by the EnvironmentVariables struct.
func Init() {

	// Attempt to load and store the environment variables. Failure here will
	// cause a panic - the server can not function if the database's environment variables are not
	// present, or could not be loaded properly.
	err := envvars.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Based on the environment variables that were loaded above, construct all of the prepared statements that
	// the database will use.
	pstatements.Construct(&envvars)

	// Construct the connection string for the database connection.
	var connString = fmt.Sprintf("%v:%v@(%v:%v)/%v?tls=skip-verify&parseTime=true", envvars.DBUsername, envvars.DBPass, envvars.DBURL, envvars.DBPort, envvars.DBName)

	// Attempt to open the connection based on the connection string above. Failure here will
	// cause a panic, as the server cannot function is the database instance is not valid.
	// The resultant database object when successful is stored in the instance variable for this package.
	db, err = sql.Open("mysql", connString)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database connection initiated successfully")
}

// ValidateAuth checks the specified database ID and token to see if they match and are valid.
func ValidateAuth(publicID string, authToken string) (id uint64, err error) {

	// Attempt to get the user's Database ID, and ban status.
	id, banned, err := getUser(publicID)
	if err != nil {
		return id, err
	}

	// Exit earlier with an error if the user is banned.
	if banned {
		return id, errors.New("User is banned")
	}

	// Prepare a statement that will fetch the expiry datetime for the specified user's auth token.
	statement, err := db.Prepare(pstatements.GetAuthExpiry)
	if err != nil {
		return id, errors.New("Internal server error: Failed to prepare statement")
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the tokens table with the specified database ID and auth token.
	// The returned row should have a single column - the expiry of datetime for the auth token.
	// An error means that either a row was not found, or there was a database error.
	var expiry time.Time
	err = statement.QueryRow(id, authToken).Scan(&expiry)
	if err != nil {
		return id, errors.New("Token is invalid")
	}

	// If the token is expired (less than [authExpiryGracePeriod] time remains until the expiry datetime), return
	// an appropriate error.
	if expiry.Sub(time.Now()) <= authExpiryGracePeriod {
		return id, errors.New("Token is expired")
	}

	return id, err
}

// GetMMR returns the current MMR for the specified user.
func GetMMR(databaseID uint64) (MMR int, err error) {

	// Prepare a statement that will fetch the MMR for the specified user.
	statement, err := db.Prepare(pstatements.GetMMR)
	if err != nil {
		return MMR, errors.New("Internal server error: Failed to prepare statement")
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the profiles table with the specified database ID.
	// The returned row should have a single column - the MMR for the user.
	// An error means that either a row was not found, or there was a database error.
	err = statement.QueryRow(databaseID).Scan(&MMR)
	if err != nil {
		return MMR, errors.New("User does not exist")
	}

	return MMR, nil
}

// CreateMatch creates a match with the two clients specified, and returns the match id.
func CreateMatch(client1DatabaseID uint64, client2DatabaseID uint64) (matchID int64, err error) {

	// Prepare a statement that will add an entry to the matches table with the specified match details.
	statement, err := db.Prepare(pstatements.CreateMatch)
	if err != nil {
		return matchID, errors.New("Internal server error: Failed to prepare statement")
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the matches table with the specified database ID's.
	// The returned value contains information about the outcome of executing the command.
	// An error means that either the specified values were invalid, or there was a database error.
	res, err := statement.Exec(client1DatabaseID, client2DatabaseID)
	if err != nil {
		return matchID, err
	}

	// Read the last insert ID from the result from the previous query - this is the match ID,
	// which is used as the return value for this function.
	matchID, err = res.LastInsertId()
	if err != nil {
		return matchID, err
	}

	return matchID, err
}

// ValidateMatch returns true if the specified match exists, and the specified client is part of it.
func ValidateMatch(databaseID uint64, matchID uint64) (valid bool, err error) {

	// Prepare a statement that will check if a match exists in the matches table with the specified match
	// ID, and the specified user is present.
	statement, err := db.Prepare(pstatements.CheckMatchValid)

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the matches table with the specified user and match ID.
	// The returned row should have a single column - the outcome (true or false) of the query.
	// An error means that either the row was not found, or there was a database error.
	var found bool
	err = statement.QueryRow(matchID, databaseID).Scan(&found)
	if err == sql.ErrNoRows {
		return false, errors.New("Invalid - either the match does not exist, or the specified client is not part of it")
	} else if err != nil {
		return false, err
	}

	return found, nil
}

// GetClientNameAndAvatar returns the displayname and avatar id for the specified user.
func GetClientNameAndAvatar(databaseID uint64) (displayname string, avatar uint8, err error) {

	// Prepare a statement that will fetch the display name for the specified user.
	statement, err := db.Prepare(pstatements.GetDisplayName)
	if err != nil {
		return displayname, 0, errors.New("Internal server error: Failed to prepare statement")
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the users table with the specified database ID.
	// The returned row should have a single column - the display name for the user.
	// An error means that either a row was not found, or there was a database error.
	err = statement.QueryRow(databaseID).Scan(&displayname)
	if err != nil {
		return displayname, 0, errors.New("User does not exist")
	}

	// Close the previous statement, so that its resources are cleared (locally and/or on the database).
	err = statement.Close()
	if err != nil {
		return displayname, 0, errors.New("Failed to close statement")
	}

	// Prepare a statement that will fetch the avatar id for the specified user.
	statement, err = db.Prepare(pstatements.GetAvatar)
	if err != nil {
		return displayname, 0, errors.New("Internal server error: Failed to prepare statement")
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the profiles table with the specified database ID.
	// The returned row should have a single column - the avatar id for the user.
	// An error means that either a row was not found, or there was a database error.
	err = statement.QueryRow(databaseID).Scan(&avatar)
	if err != nil {
		return displayname, 0, errors.New("User does not exist")
	}

	return displayname, avatar, nil
}

// SetMatchStart updates the phase + start time column for the specified match.
func SetMatchStart(matchID uint64) (err error) {

	// Prepare a statement that will update the row in the matches table with the specified match ID.
	statement, err := db.Prepare(pstatements.SetMatchStart)
	if err != nil {
		return errors.New("Internal server error: Failed to prepare statement")
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the matches table with the specified match ID.
	// The returned value is ignored, as it will not contain any data that we need.
	// An error means that either the specified values were invalid, or there was a database error.
	_, err = statement.Exec(matchID)
	if err != nil {
		return err
	}

	return err
}

// SetMatchResult updates the entire match specified with the winner, end time, and sets phase to 2 (finished).
func SetMatchResult(matchID uint64, winnerDatabaseID uint64) (err error) {

	// Prepare a statement that will update the row in the matches table with the specified match ID.
	statement, err := db.Prepare(pstatements.SetMatchResult)
	if err != nil {
		return errors.New("Internal server error: Failed to prepare statement")
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the matches table with the new match phase (2 - ended), specified match ID, and the databaseID of the winning player.
	// The returned value is ignored, as it will not contain any data that we need.
	// An error means that either the specified values were invalid, or there was a database error.
	_, err = statement.Exec(2, winnerDatabaseID, matchID)
	if err != nil {
		return err
	}

	return err
}

// getUser is a helper function that returns the database ID and ban state for the specified user
func getUser(publicID string) (databaseID uint64, banned bool, err error) {

	// Prepare a statement that will query the users table with the specified public ID.
	statement, err := db.Prepare(pstatements.GetUser)
	if err != nil {
		return databaseID, banned, errors.New("Internal server error: Failed to prepare statement")
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the profiles table with the specified public ID.
	// The returned row should have a two columns - the database ID, and the ban state (true or false) for the user.
	// An error means that either a row was not found, or there was a database error.
	err = statement.QueryRow(publicID).Scan(&databaseID, &banned)
	if err != nil {
		return databaseID, banned, errors.New("User does not exist")
	}

	return databaseID, banned, nil
}
