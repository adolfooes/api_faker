package response

import (
	"encoding/json"
	"net/http"
)

// Response is the common structure for all API responses.
type Response struct {
	Message string      `json:"message"`
	Stack   string      `json:"stack,omitempty"` // Stack can be omitted if empty
	Data    interface{} `json:"data,omitempty"`  // Data can be omitted if nil
}

// SendResponse is a helper function to send standardized API responses or mock responses
func SendResponse(w http.ResponseWriter, statusCode int, message string, stack string, data interface{}, returnOnlyMockedValue bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode) // Set the HTTP status code

	if returnOnlyMockedValue {
		// Send the mocked JSON data
		if mockData, ok := data.(string); ok {
			w.Write([]byte(mockData))
		} else {
			// If data is not a string, handle error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Invalid mock data"}`))
		}
		return
	}

	// Send the structured response
	response := Response{
		Message: message,
		Stack:   stack, // change to send only in development mode
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}
