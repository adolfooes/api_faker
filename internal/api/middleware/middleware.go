package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey = []byte("your_secret_key") // Same secret key

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

		// // Optionally: You can check for additional claims here like expiration time (exp)
		// // For example, if you want to ensure the token has not expired:
		// if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 	if ExpiresAt, ok := claims["exp"].(float64); ok {
		// 		// Compare exp with current time in Unix format if necessary
		// 		// e.g., time.Now().Unix() > int64(exp)
		// 	}
		// }

		// Token is valid, continue with the request
		next.ServeHTTP(w, r)
	})
}
