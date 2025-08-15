package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func SetupTestDatabase(path string) (*sql.DB, error) {
	os.Remove(path)
	query, err := os.ReadFile(os.Getenv("DB_SCHEMA_PATH"))
	if err != nil {
		return nil, fmt.Errorf("DB setup -> %s", err)
	}
	db, err := sql.Open("sqlite3", "./database.db")
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
