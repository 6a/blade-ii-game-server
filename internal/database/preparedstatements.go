// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package database provides an interface through which the application can interact with a database.
package database

import (
	"fmt"
	"log"
)

// PreparedStatements is a light wrapper for all the prepared statements used in this package.
type PreparedStatements struct {
	GetUser         string
	GetAuthExpiry   string
	GetMMR          string
	CreateMatch     string
	CheckMatchValid string
	GetDisplayName  string
	GetAvatar       string
	SetMatchStart   string
	SetMatchResult  string
}

// Construct constructs all the prepared statements for this PreparedStatements object.
func (p *PreparedStatements) Construct(envvars *EnvironmentVariables) {

	// Get the "id" and "banned" columns from the row in the users table with the specified public ID.
	p.GetUser = fmt.Sprintf("SELECT `id`, `banned` FROM `%v`.`%v` WHERE `public_id` = ?;", envvars.DBName, envvars.TableUsers)

	// Get the "auth_expiry" column from the row in the tokens table with the specified database ID.
	p.GetAuthExpiry = fmt.Sprintf("SELECT `auth_expiry` FROM `%v`.`%v` WHERE `id` = ? AND `auth` = ?;", envvars.DBName, envvars.TableTokens)

	// Get the "mmr" column from the row in the profiles table with the specified database ID.
	p.GetMMR = fmt.Sprintf("SELECT `mmr` FROM `%v`.`%v` WHERE `id` = ?;", envvars.DBName, envvars.TableProfiles)

	// Insert a new row into the matches table and set the "player1" and "player2" columns with the specified values.
	p.CreateMatch = fmt.Sprintf("INSERT INTO `%v`.`%v` (`player1`, `player2`) VALUES (?, ?);", envvars.DBName, envvars.TableMatches)

	// Return a row with a value of either true of false, based on whether a row exists in the matches table with the specified match ID, and where "player1"
	// or "player2" matches the specified database ID.
	p.CheckMatchValid = fmt.Sprintf("SELECT EXISTS (SELECT * FROM `%v`.`%v` WHERE `id` = ? AND `phase` = 0 AND ? IN(`player1`, `player2`));", envvars.DBName, envvars.TableMatches)

	// Get the "handle" column from the row in the users table with the specified database ID.
	p.GetDisplayName = fmt.Sprintf("SELECT `handle` FROM `%v`.`%v` WHERE `id` = ?;", envvars.DBName, envvars.TableUsers)

	// Get the "avatar" column from the row in the profiles table with the specified database ID.
	p.GetAvatar = fmt.Sprintf("SELECT `avatar` FROM `%v`.`%v` WHERE `id` = ?;", envvars.DBName, envvars.TableProfiles)

	// Update the "phase" and "start" column for the row in the matches table with the specified match ID.
	p.SetMatchStart = fmt.Sprintf("UPDATE `%v`.`%v` SET `phase` = 1, `start` = NOW() WHERE `id` = ?;", envvars.DBName, envvars.TableMatches)

	// Update the "phase", "winner", and "end" column for the row in the matches table with the specified match ID.
	p.SetMatchResult = fmt.Sprintf("UPDATE `%v`.`%v` SET `phase` = ?, `winner` = ?, `end` = NOW() WHERE `id` = ?;", envvars.DBName, envvars.TableMatches)

	log.Println("Prepared statements constructed successfully")
}
