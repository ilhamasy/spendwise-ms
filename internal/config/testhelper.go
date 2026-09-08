package config

import (
	"os"
	"testing"
)

// InitTestDB sets environment to use test database and initializes the connection.
// Must be called before any test that uses the database.
// This prevents tests from affecting the main application database.
func InitTestDB(t *testing.T) {
	t.Helper()

	// Set test database name before loading config
	os.Setenv("DB_NAME", "spendwise_test_db")

	// Reset DB to force fresh connection with test database
	DB = nil

	cfg := Load()
	InitDB(cfg)
}
