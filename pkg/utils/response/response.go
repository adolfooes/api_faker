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
		// Handle if the mock data is a string or already JSON-encoded object
		switch v := data.(type) {
		case string:
			// Write string data directly
			w.Write([]byte(v))
		case []byte:
			// Write byte slice (in case the mock data is a raw JSON byte array)
			w.Write(v)
		default:
			// Assume it's a JSON object and encode it
			if err := json.NewEncoder(w).Encode(data); err != nil {
				w.Write([]byte(`{"error": "Invalid mock data"}`))
			}
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
