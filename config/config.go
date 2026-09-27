package config

import (
	"os"
	"strconv"
	"time"
)

const defaultWebhookTimeout = 10 * time.Second

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

// GetWebhookTimeout returns the outbound webhook request timeout, configurable via
// WEBHOOK_TIMEOUT_SECONDS. Defaults to 10s. No automatic retry is performed on timeout.
func GetWebhookTimeout() time.Duration {
	raw := os.Getenv("WEBHOOK_TIMEOUT_SECONDS")
	if raw == "" {
		return defaultWebhookTimeout
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return defaultWebhookTimeout
	}
	return time.Duration(seconds) * time.Second
}
