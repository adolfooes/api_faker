package handler

import (
	"encoding/json"
	"net/http"
)

// MockData represents a generic response for mocking
var MockData = map[string]interface{}{
	"status":  "success",
	"message": "This is mock data",
}

// GetMock handles the retrieval of mock data
func GetMock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MockData)
}

// CreateMock handles the creation of mock data (for testing)
func CreateMock(w http.ResponseWriter, r *http.Request) {
	// Logic to simulate creation
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Mock data created"))
}

// UpdateMock handles the updating of mock data (for testing)
func UpdateMock(w http.ResponseWriter, r *http.Request) {
	// Logic to simulate update
	w.Write([]byte("Mock data updated"))
}

// DeleteMock handles the deletion of mock data (for testing)
func DeleteMock(w http.ResponseWriter, r *http.Request) {
	// Logic to simulate deletion
	w.Write([]byte("Mock data deleted"))
}
