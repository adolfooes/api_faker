package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// Account represents the account structure
type Account struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// hashPassword hashes the given password using bcrypt
func hashPassword(password string) (string, error) {
	// Generate hashed password with bcrypt, the cost parameter 14 is usually a good default
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CreateAccountHandler handles the creation of a new account
func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var account Account

	// Decode the request body into the account struct
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Hash the password before storing it
	hashedPassword, err := hashPassword(account.Password)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to hash password", err.Error(), nil, false)
		return
	}
	account.Password = hashedPassword

	// Insert the new account into the database using the crud package
	columns := []string{"email", "password"}
	values := []interface{}{account.Email, account.Password}
	createdAccount, err := crud.Create("account", columns, values) // Fetching the created account object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to create account", err.Error(), nil, false)
		return
	}

	// Remove password from the response
	if createdAccount != nil {
		createdAccount["password"] = nil
	}

	response.SendResponse(w, http.StatusCreated, "Account created successfully", "", createdAccount, false)
}

// GetAllAccountsHandler retrieves all accounts from the database
func GetAllAccountsHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all accounts
	results, err := crud.List("account", filters)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve accounts", err.Error(), nil, false)
		return
	}

	// Remove password from the response
	for _, result := range results {
		result["password"] = nil
	}

	response.SendResponse(w, http.StatusOK, "Accounts retrieved successfully", "", results, false)
}

// GetAccountHandler retrieves a single account by ID from the database
func GetAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the {id} from the URL path using Gorilla Mux's Vars
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		// If the id is missing in the path, return a bad request
		response.SendResponse(w, http.StatusBadRequest, "ID is missing in the request", "", nil, false)
		return
	}

	// Convert the id from string to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	// Fetch the account from the database using the crud.Read function
	result, err := crud.Read("account", id)
	if err != nil {
		response.SendResponse(w, http.StatusNotFound, "Failed to retrieve account", err.Error(), nil, false)
		return
	}

	// Remove password from the response
	if result != nil {
		result["password"] = nil
	}

	response.SendResponse(w, http.StatusOK, "Account retrieved successfully", "", result, false)
}

// UpdateAccountHandler handles updating an existing account
func UpdateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var account Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// If the password is being updated, hash the new password
	if account.Password != "" {
		hashedPassword, err := hashPassword(account.Password)
		if err != nil {
			response.SendResponse(w, http.StatusInternalServerError, "Failed to hash password", err.Error(), nil, false)
			return
		}
		account.Password = hashedPassword
	}

	// Update the account in the database using the crud package
	updates := map[string]interface{}{
		"email":    account.Email,
		"password": account.Password,
	}

	updatedAccount, err := crud.Update("account", account.ID, updates) // Fetching the updated account object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to update account", err.Error(), nil, false)
		return
	}

	// Remove password from the response
	if updatedAccount != nil {
		updatedAccount["password"] = nil
	}

	response.SendResponse(w, http.StatusOK, "Account updated successfully", "", updatedAccount, false)
}

// DeleteAccountHandler handles deleting an account
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	err = crud.Delete("account", id)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to delete account", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "Account deleted successfully", "", nil, false)
}
