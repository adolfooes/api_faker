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
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Insert the new URL HTTP status into the database using the crud package
	columns := []string{"url_id", "http_status", "percentage"}
	values := []interface{}{status.URLID, status.HTTPStatus, status.Percentage}
	createdStatus, err := crud.Create("url_http_status", columns, values) // Fetch the created object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to create HTTP status", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusCreated, "HTTP status created successfully", "", createdStatus, false)
}

// GetAllURLHTTPStatusesHandler retrieves all HTTP statuses from the database
func GetAllURLHTTPStatusesHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all HTTP statuses
	results, err := crud.List("url_http_status", filters)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve HTTP statuses", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "HTTP statuses retrieved successfully", "", results, false)
}

// UpdateURLHTTPStatusHandler handles updating an existing HTTP status
func UpdateURLHTTPStatusHandler(w http.ResponseWriter, r *http.Request) {
	var status URLHTTPStatus
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Update the HTTP status in the database using the crud package
	updates := map[string]interface{}{
		"url_id":      status.URLID,
		"http_status": status.HTTPStatus,
		"percentage":  status.Percentage,
	}
	updatedStatus, err := crud.Update("url_http_status", status.ID, updates) // Fetch the updated object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to update HTTP status", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "HTTP status updated successfully", "", updatedStatus, false)
}

// DeleteURLHTTPStatusHandler handles deleting an HTTP status
func DeleteURLHTTPStatusHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	err = crud.Delete("url_http_status", id)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to delete HTTP status", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "HTTP status deleted successfully", "", nil, false)
}
