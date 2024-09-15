package config

import (
	"os"
)

// GetDatabaseConnectionString returns the database connection string from an environment variable
func GetDatabaseConnectionString() string {

	connStr := os.Getenv("FAKER_DATABASE_URL")

	if connStr == "" {
		connStr = "postgres://postgres:dev123@db:5432/api_faker_dev?sslmode=disable" // Replace with your default values
	}

	return connStr
}

func GetJWTSecretKey() []byte {
	return []byte(os.Getenv("JWT_SECRET_KEY"))
}
