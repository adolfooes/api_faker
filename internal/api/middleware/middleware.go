package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/adolfooes/api_faker/config"
)

var jwtSecretKey = []byte("your_secret_key") // Same secret key

// Key to use when setting the account ID in context
type contextKey string

// JWTMiddleware checks for the token in the Authorization header
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Trim the "Bearer " prefix from the Authorization header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Define a struct to store claims
		claims := jwt.MapClaims{}

		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Ensure that the method used for signing is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrUseLastResponse
			}
			return jwtSecretKey, nil
		})

		// Handle invalid tokens or errors
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		var accountID string
		if id, ok := claims["account_id"].(float64); ok {
			accountID = strconv.FormatInt(int64(id), 10)
		} else {
			http.Error(w, "Account ID not found in token or not a float64", http.StatusUnauthorized)
			return
		}

		// Inject the account ID into the request's context
		ctx := context.WithValue(r.Context(), config.JWTAccountIDKey, accountID)

		// Continue the request with the new context
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
