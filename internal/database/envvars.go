package database

import (
	"errors"
	"log"
	"os"
)

// EnvironmentVariables is a light wrapper for the environmental variables required for the database package
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

// Load attempts to read in all the required environment variables
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
