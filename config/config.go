package config

import (
	"log"
	"os"
)

// GetDatabaseConnectionString returns the database connection string from an environment variable
func GetDatabaseConnectionString() string {
	// Check if FAKER_DATABASE_URL is set in environment variables
	connStr := os.Getenv("FAKER_DATABASE_URL")

	log.Printf("Database connection string a: %s", connStr)

	// If not set, return a default or fallback connection string
	if connStr == "" {
		connStr = "postgres://postgres:dev123@db:5432/api_faker_dev?sslmode=disable" // Replace with your default values
	}

	log.Printf("Database connection string a: %s", connStr)

	return connStr
}
