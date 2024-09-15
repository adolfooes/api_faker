package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/adolfooes/api_faker/config"
	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Credentials represents the user's login credentials
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Claims represents the structure of the JWT claims, including account_id
type Claims struct {
	Email     string `json:"email"`
	AccountID int64  `json:"account_id"`
	jwt.RegisteredClaims
}

// Email validation function
func validateEmailFormat(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// Password validation function (non-empty check)
func validatePasswordNotEmpty(password string) error {
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

// LoginHandler handles user login and generates JWT token
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Validate email format
	if err := validateEmailFormat(creds.Email); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid email format", err.Error(), nil, false)
		return
	}

	// Validate password is not empty
	if err := validatePasswordNotEmpty(creds.Password); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid credentials", err.Error(), nil, false)
		return
	}

	// Search for the account using the email (case-insensitive)
	filter := map[string]interface{}{"email": strings.ToLower(creds.Email)}
	results, err := crud.List("account", filter) // Alternatively, modify Read to accept filters
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Error searching for account", err.Error(), nil, false)
		return
	}

	if len(results) == 0 {
		// No account found with that email
		response.SendResponse(w, http.StatusUnauthorized, "Invalid credentials", "", nil, false)
		return
	}

	// Since we're querying by email, only one result is expected
	account := results[0]

	// Compare the provided password with the hashed password in the database
	storedPassword := account["password"].(string)
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password))
	if err != nil {
		// Password does not match
		response.SendResponse(w, http.StatusUnauthorized, "Invalid credentials", "", nil, false)
		return
	}

	// Get the account ID from the account object
	accountID := account["id"].(int64)

	// Set JWT expiration time
	expirationTime := time.Now().Add(1 * time.Hour)

	// Create JWT claims, including the user's email, account_id, and expiration time
	claims := &Claims{
		Email:     creds.Email,
		AccountID: accountID, // Add account ID to the claims
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Generate the token with claims and sign it with the secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.GetJWTSecretKey())
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to generate token", err.Error(), nil, false)
		return
	}

	// Send the JWT token to the client
	response.SendResponse(w, http.StatusOK, "Login successful", "", map[string]string{"token": tokenString}, false)
}
