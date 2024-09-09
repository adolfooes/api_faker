package response

import (
	"encoding/json"
	"net/http"
)

// Response is the common structure for all API responses, including optional pagination info
type Response struct {
	Message    string      `json:"message"`
	Stack      string      `json:"stack,omitempty"`      // Stack can be omitted if empty
	Data       interface{} `json:"data,omitempty"`       // Data can be omitted if nil
	Pagination *Pagination `json:"pagination,omitempty"` // Optional pagination info
}

// Pagination contains the pagination metadata
type Pagination struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalItems  int `json:"totalItems"`
	TotalPages  int `json:"totalPages"`
}

// SendResponse is a helper function to send standardized API responses
func SendResponse(w http.ResponseWriter, message string, stack string, data interface{}, pagination *Pagination) {
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Message:    message,
		Stack:      stack,
		Data:       data,
		Pagination: pagination,
	}
	json.NewEncoder(w).Encode(response)
}
