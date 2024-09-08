package handler

import (
	"encoding/json"
	"net/http"
)

// GetAccounts handles the request to retrieve all accounts
func GetAccounts(w http.ResponseWriter, r *http.Request) {
	accounts := []map[string]interface{}{
		{"id": 1, "username": "user1"},
		{"id": 2, "username": "user2"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// GetAccount handles the request to retrieve a single account by ID
func GetAccount(w http.ResponseWriter, r *http.Request) {
	// Example of retrieving account by ID
	account := map[string]interface{}{"id": 1, "username": "user1"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

// CreateAccount handles the creation of a new account
func CreateAccount(w http.ResponseWriter, r *http.Request) {
	// Logic to create a new account
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Account created"))
}

// UpdateAccount handles the updating of an existing account by ID
func UpdateAccount(w http.ResponseWriter, r *http.Request) {
	// Logic to update an account by ID
	w.Write([]byte("Account updated"))
}

// DeleteAccount handles the deletion of an account by ID
func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	// Logic to delete an account by ID
	w.Write([]byte("Account deleted"))
}
