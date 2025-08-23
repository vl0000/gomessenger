package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func SetupTestDatabase(path string) (*sql.DB, error) {
	query, err := os.ReadFile(os.Getenv("DB_SCHEMA_PATH"))
	if err != nil {
		return nil, fmt.Errorf("DB setup -> %s", err)
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("DB setup -> %s", err)
	}

	if _, err = db.Exec(string(query)); err != nil {
		db.Close()
		return nil, fmt.Errorf("DB setup -> %s", err)
	}

	log.Println("Database setup")
	return db, nil

}
