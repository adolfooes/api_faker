package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var db *sql.DB

// InitDB initializes the database connection using a provided connection string
func InitDB(connStr string) {
	var err error

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening the database: %v", err)
	}

	// Verify the connection with the database
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Database connection established successfully")
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("The database has not been initialized")
	}
	return db
}
