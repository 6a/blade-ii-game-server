// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package database provides an interface through which the application can interact with a database.
package database

import (
	"errors"
	"log"
	"os"
)

// EnvironmentVariables is a light wrapper for the environment variables required by the database package.
type EnvironmentVariables struct {
	DBUsername    string
	DBPass        string
	DBURL         string
	DBPort        string
	DBName        string
	TableUsers    string
	TableProfiles string
	TableMatches  string
	TableTokens   string
}

// Load attempts to read in all the required environment variables.
func (ev *EnvironmentVariables) Load() error {
	ev.DBUsername = os.Getenv("db_user")
	ev.DBPass = os.Getenv("db_pass")
	ev.DBURL = os.Getenv("db_url")
	ev.DBPort = os.Getenv("db_port")
	ev.DBName = os.Getenv("db_name")
	ev.TableUsers = os.Getenv("db_table_users")
	ev.TableProfiles = os.Getenv("db_table_profiles")
	ev.TableMatches = os.Getenv("db_table_matches")
	ev.TableTokens = os.Getenv("db_table_tokens")

	// Check all the loaded values - empty strings suggest that either the environment variable
	// did not exist, or exists but has no value (or was an empty string etc.). If any variable
	// was not loaded with a valid value, return an error.

	if ev.DBUsername == "" {
		return errors.New("Environment variable [db_user] was not set, or is empty")
	}

	if ev.DBPass == "" {
		return errors.New("Environment variable [db_pass] was not set, or is empty")
	}

	if ev.DBURL == "" {
		return errors.New("Environment variable [db_url] was not set, or is empty")
	}

	if ev.DBPort == "" {
		return errors.New("Environment variable [db_port] was not set, or is empty")
	}

	if ev.DBName == "" {
		return errors.New("Environment variable [db_name] was not set, or is empty")
	}

	if ev.TableUsers == "" {
		return errors.New("Environment variable [db_table_users] was not set, or is empty")
	}

	if ev.TableProfiles == "" {
		return errors.New("Environment variable [db_table_profiles] was not set, or is empty")
	}

	if ev.TableMatches == "" {
		return errors.New("Environment variable [db_table_matches] was not set, or is empty")
	}

	if ev.TableTokens == "" {
		return errors.New("Environment variable [db_table_tokens] was not set, or is empty")
	}

	log.Println("Environment variables loaded successfully")

	return nil
}
