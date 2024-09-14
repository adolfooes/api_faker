package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

func RunMigrations(connStr string) {
	log.Println("Starting migrations...")

	// Initialize the migration
	m, err := migrate.New(
		"file:///migrations", // Path to your migration files
		connStr,
	)

	if err != nil {
		log.Fatalf("Failed to initialize migration: %v", err)
	}

	// Apply all migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations completed successfully.")
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("The database has not been initialized")
	}
	return db
}
