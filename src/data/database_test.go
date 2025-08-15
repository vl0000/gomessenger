package data_test

import (
	"os"
	"testing"

	"github.com/vl0000/gomessenger/data"
)

func TestSetupTestDB(t *testing.T) {
	t.Run("Test db can setup", func(t *testing.T) {
		os.Setenv("DB_SCHEMA_PATH", "./data/database.sql")
		data.SetupTestDatabase("./testdb.db")
		os.Remove("./testdb.db")
	})
}
