package config

import "os"

// GetDatabaseConnectionString returns the database connection string from an environment variable
func GetDatabaseConnectionString() string {
	// Check if DATABASE_URL is set in environment variables
	connStr := os.Getenv("FAKER_DATABASE_URL")

	// If not set, return a default or fallback connection string
	if connStr == "" {
		connStr = "postgres://dev:dev123@localhost/faker_api?sslmode=disable" // Replace with your default values
	}

	return connStr
}
