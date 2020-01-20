package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexedwards/argon2id"
	_ "github.com/go-sql-driver/mysql" // mysql driver
)

var db *sql.DB

const passwordCheckConstantTimeMin = time.Millisecond * 1500

var argonParams = argon2id.Params{
	Memory:      32 * 1024,
	Iterations:  3,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   30,
}

var dbuser = os.Getenv("db_user")
var dbpass = os.Getenv("db_pass")
var dburl = os.Getenv("db_url")
var dbport = os.Getenv("db_port")
var dbname = os.Getenv("db_name")
var dbtableUsers = os.Getenv("db_table_users")
var dbtableProfiles = os.Getenv("db_table_profiles")
var dbtableMatches = os.Getenv("db_table_matches")
var dbtableTokens = os.Getenv("db_table_tokens")

var psCreateAccount = fmt.Sprintf("INSERT INTO `%v`.`%v` (`public_id`, `handle`, `email`, `salted_hash`) VALUES (?, ?, ?, ?);", dbname, dbtableUsers)
var psCreateTokenRowWithEmailToken = fmt.Sprintf("INSERT INTO `%v`.`%v` (`id`, `email_confirmation`, `email_confirmation_expiry`) VALUES (LAST_INSERT_ID(), ?, DATE_ADD(NOW(), INTERVAL ? HOUR));", dbname, dbtableTokens)
var psAddTokenWithReplacers = fmt.Sprintf("UPDATE `%v`.`%v` SET `repl_1` = ?, `repl_2` = DATE_ADD(NOW(), INTERVAL ? HOUR) WHERE `id` = ?;", dbname, dbtableTokens)

var psCheckName = fmt.Sprintf("SELECT EXISTS(SELECT * FROM `%v`.`%v` WHERE `handle` = ?);", dbname, dbtableUsers)
var psCheckAuth = fmt.Sprintf("SELECT `salted_hash`, `banned` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)
var psGetIDs = fmt.Sprintf("SELECT `id`, `public_id` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)

var connString = fmt.Sprintf("%v:%v@(%v:%v)/%v?tls=skip-verify", dbuser, dbpass, dburl, dbport, dbname)

// Init should be called at the start of the function to open a connection to the database
func Init() {
	mysql, err := sql.Open("mysql", connString)
	if err != nil {
		log.Fatal(err)
	}

	db = mysql
}

// ValidateAuth checks the specified user ID and token to see if they match and are valid
func ValidateAuth(uid string, token string) (err error) {
	return err
}

// GetMMR returns the current MMR for the specified user
func GetMMR(uid string) (mmr int, err error) {
	return 0, err
}
