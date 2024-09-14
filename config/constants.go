package config

// Define a custom type for context keys to avoid potential conflicts
type ContextKey string

// Constant for the account ID context key
const JWTAccountIDKey ContextKey = "account_id"
