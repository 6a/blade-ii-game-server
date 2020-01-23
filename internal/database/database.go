package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

const authExpiryGracePeriod = time.Minute * 10

var db *sql.DB

var dbuser = os.Getenv("db_user")
var dbpass = os.Getenv("db_pass")
var dburl = os.Getenv("db_url")
var dbport = os.Getenv("db_port")
var dbname = os.Getenv("db_name")
var dbtableUsers = os.Getenv("db_table_users")
var dbtableProfiles = os.Getenv("db_table_profiles")
var dbtableMatches = os.Getenv("db_table_matches")
var dbtableTokens = os.Getenv("db_table_tokens")

var psGetUser = fmt.Sprintf("SELECT `id`, `banned` FROM `%v`.`%v` WHERE `public_id` = ?;", dbname, dbtableUsers)
var psGetAuthExpiry = fmt.Sprintf("SELECT `auth_expiry` FROM `%v`.`%v` WHERE `id` = ? AND `auth` = ?;", dbname, dbtableTokens)
var psGetMMR = fmt.Sprintf("SELECT `mmr` FROM `%v`.`%v` WHERE `id` = ?;", dbname, dbtableProfiles)

// var psCreateAccount = fmt.Sprintf("INSERT INTO `%v`.`%v` (`public_id`, `handle`, `email`, `salted_hash`) VALUES (?, ?, ?, ?);", dbname, dbtableUsers)
// var psCreateTokenRowWithEmailToken = fmt.Sprintf("INSERT INTO `%v`.`%v` (`id`, `email_confirmation`, `email_confirmation_expiry`) VALUES (LAST_INSERT_ID(), ?, DATE_ADD(NOW(), INTERVAL ? HOUR));", dbname, dbtableTokens)
// var psAddTokenWithReplacers = fmt.Sprintf("UPDATE `%v`.`%v` SET `repl_1` = ?, `repl_2` = DATE_ADD(NOW(), INTERVAL ? HOUR) WHERE `id` = ?;", dbname, dbtableTokens)

// var psCheckName = fmt.Sprintf("SELECT EXISTS(SELECT * FROM `%v`.`%v` WHERE `handle` = ?);", dbname, dbtableUsers)
// var psCheckAuth = fmt.Sprintf("SELECT `salted_hash`, `banned` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)
// var psGetIDs = fmt.Sprintf("SELECT `id`, `public_id` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)

var connString = fmt.Sprintf("%v:%v@(%v:%v)/%v?tls=skip-verify&parseTime=true", dbuser, dbpass, dburl, dbport, dbname)

// Init should be called at the start of the function to open a connection to the database
func Init() {
	mysql, err := sql.Open("mysql", connString)
	if err != nil {
		log.Fatal(err)
	}

	db = mysql
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

	statement, err := db.Prepare(psGetAuthExpiry)
	defer statement.Close()
	if err != nil {
		return id, errors.New("Internal server error: Failed to prepare statement")
	}

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
	statement, err := db.Prepare(psGetMMR)
	defer statement.Close()
	if err != nil {
		return mmr, errors.New("Internal server error: Failed to prepare statement")
	}

	err = statement.QueryRow(id).Scan(&mmr)
	if err != nil {
		return mmr, errors.New("User does not exist")
	}

	return mmr, nil
}

func getUser(pid string) (id uint64, banned bool, err error) {
	statement, err := db.Prepare(psGetUser)
	defer statement.Close()
	if err != nil {
		return id, banned, errors.New("Internal server error: Failed to prepare statement")
	}

	err = statement.QueryRow(pid).Scan(&id, &banned)
	if err != nil {
		return id, banned, errors.New("User does not exist")
	}

	return id, banned, nil
}
