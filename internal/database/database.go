package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

const authExpiryGracePeriod = time.Minute * 10

var db *sql.DB
var envvars EnvironmentVariables
var pstatements PreparedStatements

// Init should be called at the start of the function to open a connection to the database
func Init() {
	err := envvars.Load()
	if err != nil {
		log.Fatal(err)
	}

	pstatements.Construct(&envvars)

	var connString = fmt.Sprintf("%v:%v@(%v:%v)/%v?tls=skip-verify&parseTime=true", envvars.User, envvars.Pass, envvars.URL, envvars.Port, envvars.Name)
	mysql, err := sql.Open("mysql", connString)
	if err != nil {
		log.Fatal(err)
	}

	db = mysql

	log.Println("Database connection initiated successfully")
}

// ValidateAuth checks the specified user ID and token to see if they match and are valid
func ValidateAuth(pid string, token string) (id uint64, err error) {
	id, banned, err := getUser(pid)
	if err != nil {
		return id, err
	}

	if banned {
		return id, errors.New("User is banned")
	}

	statement, err := db.Prepare(pstatements.GetAuthExpiry)
	if err != nil {
		return id, errors.New("Internal server error: Failed to prepare statement")
	}

	defer statement.Close()

	var expiry time.Time
	err = statement.QueryRow(id, token).Scan(&expiry)
	if err != nil {
		return id, errors.New("Token is invalid")
	}

	if expiry.Sub(time.Now()) <= authExpiryGracePeriod {
		return id, errors.New("Token is expired")
	}

	return id, err
}

// GetMMR returns the current MMR for the specified user
func GetMMR(id uint64) (mmr int, err error) {
	statement, err := db.Prepare(pstatements.GetMMR)
	if err != nil {
		return mmr, errors.New("Internal server error: Failed to prepare statement")
	}

	defer statement.Close()

	err = statement.QueryRow(id).Scan(&mmr)
	if err != nil {
		return mmr, errors.New("User does not exist")
	}

	return mmr, nil
}

// CreateMatch creates a match with the two clients specified, and returns the match id
func CreateMatch(client1ID uint64, client2ID uint64) (id int64, err error) {
	statement, err := db.Prepare(pstatements.CreateMatch)
	if err != nil {
		return id, errors.New("Internal server error: Failed to prepare statement")
	}

	defer statement.Close()

	res, err := statement.Exec(client1ID, client2ID)
	if err != nil {
		return id, err
	}

	id, err = res.LastInsertId()
	if err != nil {
		return id, err
	}

	return id, err
}

// ValidateMatch returns true if the specified match exists, and the specified client is part of it
func ValidateMatch(clientID uint64, matchID uint64) (valid bool, err error) {
	statement, err := db.Prepare(pstatements.CheckMatchValid)

	defer statement.Close()

	var found bool
	err = statement.QueryRow(matchID, clientID, clientID).Scan(&found)
	if err == sql.ErrNoRows {
		return false, errors.New("Invalid - either the match does not exist, or the specified client is not part of it")
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func getUser(pid string) (id uint64, banned bool, err error) {
	statement, err := db.Prepare(pstatements.GetUser)
	if err != nil {
		return id, banned, errors.New("Internal server error: Failed to prepare statement")
	}

	defer statement.Close()

	err = statement.QueryRow(pid).Scan(&id, &banned)
	if err != nil {
		return id, banned, errors.New("User does not exist")
	}

	return id, banned, nil
}
