package database

import (
	"fmt"
	"log"
)

// PreparedStatements is a light wrapper for all the prepared statements used in this package
type PreparedStatements struct {
	GetUser           string
	GetAuthExpiry     string
	GetMMR            string
	CreateMatch       string
	CheckMatchValid   string
	GetDisplayName    string
	SetMatchPhase     string
	RecordMatchResult string
}

// Construct constructs all the prepared statements for this PreparedStatements object
func (p *PreparedStatements) Construct(envvars *EnvironmentVariables) {
	p.GetUser = fmt.Sprintf("SELECT `id`, `banned` FROM `%v`.`%v` WHERE `public_id` = ?;", envvars.Name, envvars.TableUsers)
	p.GetAuthExpiry = fmt.Sprintf("SELECT `auth_expiry` FROM `%v`.`%v` WHERE `id` = ? AND `auth` = ?;", envvars.Name, envvars.TableTokens)
	p.GetMMR = fmt.Sprintf("SELECT `mmr` FROM `%v`.`%v` WHERE `id` = ?;", envvars.Name, envvars.TableProfiles)
	p.CreateMatch = fmt.Sprintf("INSERT INTO `%v`.`%v` (`player1`, `player2`) VALUES (?, ?);", envvars.Name, envvars.TableMatches)
	p.CheckMatchValid = fmt.Sprintf("SELECT EXISTS (SELECT * FROM `%v`.`%v` WHERE `id` = ? AND `phase` = 0 AND (`player1` = ? OR `player2` = ?));", envvars.Name, envvars.TableMatches)
	p.GetDisplayName = fmt.Sprintf("SELECT `handle` FROM `%v`.`%v` WHERE `id` = ?;", envvars.Name, envvars.TableUsers)
	p.SetMatchPhase = fmt.Sprintf("UPDATE `%v`.`%v` SET `phase` = ? WHERE `id` = ?;", envvars.Name, envvars.TableMatches)
	p.RecordMatchResult = fmt.Sprintf("UPDATE `%v`.`%v` SET `state` = 2 WHERE `id` = ?;", envvars.Name, envvars.TableMatches)

	log.Println("Prepared statements constructed successfully")
}

// var psCreateTokenRowWithEmailToken = fmt.Sprintf("INSERT INTO `%v`.`%v` (`id`, `email_confirmation`, `email_confirmation_expiry`) VALUES (LAST_INSERT_ID(), ?, DATE_ADD(NOW(), INTERVAL ? HOUR));", dbname, dbtableTokens)
// var psAddTokenWithReplacers = fmt.Sprintf("UPDATE `%v`.`%v` SET `repl_1` = ?, `repl_2` = DATE_ADD(NOW(), INTERVAL ? HOUR) WHERE `id` = ?;", dbname, dbtableTokens)

// var psCheckName = fmt.Sprintf("SELECT EXISTS(SELECT * FROM `%v`.`%v` WHERE `handle` = ?);", dbname, dbtableUsers)
// var psCheckAuth = fmt.Sprintf("SELECT `salted_hash`, `banned` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)
// var psGetIDs = fmt.Sprintf("SELECT `id`, `public_id` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)
