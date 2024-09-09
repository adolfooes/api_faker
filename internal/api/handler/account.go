package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
)

// Account represents the account structure
type Account struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateAccountHandler handles the creation of a new account
func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var account Account

	// Decode the request body into the account struct
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Insert the new account into the database using the crud package
	columns := []string{"email", "password"}
	values := []interface{}{account.Email, account.Password}
	err = crud.Create("account", columns, values)
	if err != nil {
		response.SendResponse(w, "Failed to create account", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Account created successfully", "", account, nil)
}

// GetAllAccountsHandler retrieves all accounts from the database
func GetAllAccountsHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all accounts
	results, err := crud.List("account", filters)
	if err != nil {
		response.SendResponse(w, "Failed to retrieve accounts", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Accounts retrieved successfully", "", results, nil)
}

// GetAccountHandler retrieves a single account by ID from the database
func GetAccountHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, "Invalid ID parameter", err.Error(), nil, nil)
		return
	}

	result, err := crud.Read("account", id)
	if err != nil {
		response.SendResponse(w, "Failed to retrieve account", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Account retrieved successfully", "", result, nil)
}

// UpdateAccountHandler handles updating an existing account
func UpdateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var account Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Update the account in the database using the crud package
	updates := map[string]interface{}{
		"email":    account.Email,
		"password": account.Password,
	}
	err = crud.Update("account", account.ID, updates)
	if err != nil {
		response.SendResponse(w, "Failed to update account", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Account updated successfully", "", account, nil)
}

// DeleteAccountHandler handles deleting an account
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, "Invalid ID parameter", err.Error(), nil, nil)
		return
	}

	err = crud.Delete("account", id)
	if err != nil {
		response.SendResponse(w, "Failed to delete account", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Account deleted successfully", "", nil, nil)
}
