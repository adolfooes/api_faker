package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
)

// URLHTTPStatus represents a structure for an HTTP status associated with a URL
type URLHTTPStatus struct {
	ID         int `json:"id"`
	URLID      int `json:"url_id"`
	HTTPStatus int `json:"http_status"`
	Percentage int `json:"percentage"`
}

// CreateURLHTTPStatusHandler handles creating a new HTTP status for a URL
func CreateURLHTTPStatusHandler(w http.ResponseWriter, r *http.Request) {
	var status URLHTTPStatus

	// Decode the request body into the URL HTTP status struct
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Insert the new URL HTTP status into the database using the crud package
	columns := []string{"url_id", "http_status", "percentage"}
	values := []interface{}{status.URLID, status.HTTPStatus, status.Percentage}
	err = crud.Create("url_http_status", columns, values)
	if err != nil {
		response.SendResponse(w, "Failed to create HTTP status", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "HTTP status created successfully", "", status, nil)
}

// GetAllURLHTTPStatusesHandler retrieves all HTTP statuses from the database
func GetAllURLHTTPStatusesHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all HTTP statuses
	results, err := crud.List("url_http_status", filters)
	if err != nil {
		response.SendResponse(w, "Failed to retrieve HTTP statuses", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "HTTP statuses retrieved successfully", "", results, nil)
}

// UpdateURLHTTPStatusHandler handles updating an existing HTTP status
func UpdateURLHTTPStatusHandler(w http.ResponseWriter, r *http.Request) {
	var status URLHTTPStatus
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Update the HTTP status in the database using the crud package
	updates := map[string]interface{}{
		"url_id":      status.URLID,
		"http_status": status.HTTPStatus,
		"percentage":  status.Percentage,
	}
	err = crud.Update("url_http_status", status.ID, updates)
	if err != nil {
		response.SendResponse(w, "Failed to update HTTP status", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "HTTP status updated successfully", "", status, nil)
}

// DeleteURLHTTPStatusHandler handles deleting an HTTP status
func DeleteURLHTTPStatusHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, "Invalid ID parameter", err.Error(), nil, nil)
		return
	}

	err = crud.Delete("url_http_status", id)
	if err != nil {
		response.SendResponse(w, "Failed to delete HTTP status", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "HTTP status deleted successfully", "", nil, nil)
}
