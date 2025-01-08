package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite" // Import the sqlite3 driver
)

type Service struct {
	DB *sql.DB
}

func Conect(DBFile string) (*Service, error) {
	var install bool

	if _, err := os.Stat(DBFile); os.IsNotExist(err) {
		fmt.Println(DBFile)
		install = true
	}

	// Open the database
	db, err := sql.Open("sqlite", DBFile)
	if err != nil {
		return nil, err
	}

	// Create the table if it doesn't exist
	if install {
		sqlFile, err := os.ReadFile("./migration/create_table.sql")
		if err != nil {
			return nil, fmt.Errorf("failed to read SQL file: %v", err)
		}

		_, err = db.Exec(string(sqlFile))
		if err != nil {
			return nil, fmt.Errorf("failed to execute SQL queries: %v", err)
		}
	}

	// Create the service
	s := &Service{
		DB: db,
	}
	log.Printf("Connected to the database: %s", DBFile)
	return s, nil
}
