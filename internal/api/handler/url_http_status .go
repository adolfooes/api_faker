package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/config"
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

func validateRequiredURLHTTPStatusFields(status URLHTTPStatus) error {
	if status.URLID == 0 {
		return fmt.Errorf("url_id is required")
	}
	if status.HTTPStatus == 0 {
		return fmt.Errorf("http_status is required")
	}
	if status.Percentage < 0 || status.Percentage > 100 {
		return fmt.Errorf("percentage must be between 0 and 100")
	}
	return nil
}

func validateHTTPStatusCode(httpStatus int) error {
	if httpStatus < 100 || httpStatus > 599 {
		return fmt.Errorf("invalid HTTP status code: %d", httpStatus)
	}
	return nil
}

func validatePercentageDistribution(urlID int, newPercentage int) error {
	filters := map[string]interface{}{
		"url_id": urlID,
	}
	urlHTTPStatuses, err := crud.List("url_http_status", filters)
	if err != nil {
		return fmt.Errorf("error fetching existing HTTP statuses for URL: %v", err)
	}

	totalPercentage := 0
	for _, status := range urlHTTPStatuses {
		totalPercentage += int(status["percentage"].(int64))
	}

	if totalPercentage+newPercentage > 100 {
		return fmt.Errorf("total percentage exceeds 100%%")
	}

	return nil
}

func authorizeURLOwnership(urlID int, ownerID int64) error {
	filters := map[string]interface{}{
		"id": urlID,
	}
	urlConfigs, err := crud.List("url_config", filters)
	if err != nil || len(urlConfigs) == 0 {
		return fmt.Errorf("URL not found")
	}
	urlConfig := urlConfigs[0]

	if int64(urlConfig["owner_id"].(int)) != ownerID {
		return fmt.Errorf("you are not authorized to modify this URL")
	}
	return nil
}

func CreateURLHTTPStatusHandler(w http.ResponseWriter, r *http.Request) {
	var status URLHTTPStatus

	// Decode the request body into the URL HTTP status struct
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Validate required fields
	if err := validateRequiredURLHTTPStatusFields(status); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Validation failed", err.Error(), nil, false)
		return
	}

	// Validate HTTP status code
	if err := validateHTTPStatusCode(status.HTTPStatus); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid HTTP status code", err.Error(), nil, false)
		return
	}

	// Extract the owner ID from the context (injected by the JWT middleware)
	ownerIDStr, ok := r.Context().Value(config.JWTAccountIDKey).(string)
	if !ok {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: Owner ID not found", "", nil, false)
		return
	}
	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid Owner ID format", "", nil, false)
		return
	}

	// Authorize ownership of the URL
	if err := authorizeURLOwnership(status.URLID, ownerID); err != nil {
		response.SendResponse(w, http.StatusUnauthorized, err.Error(), "", nil, false)
		return
	}

	// Validate percentage distribution
	if err := validatePercentageDistribution(status.URLID, status.Percentage); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Percentage validation failed", err.Error(), nil, false)
		return
	}

	// Insert the new URL HTTP status into the database
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

func UpdateURLHTTPStatusHandler(w http.ResponseWriter, r *http.Request) {
	var status URLHTTPStatus
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Validate required fields
	if err := validateRequiredURLHTTPStatusFields(status); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Validation failed", err.Error(), nil, false)
		return
	}

	// Validate HTTP status code
	if err := validateHTTPStatusCode(status.HTTPStatus); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid HTTP status code", err.Error(), nil, false)
		return
	}

	// Extract the owner ID from the context (injected by the JWT middleware)
	ownerIDStr, ok := r.Context().Value(config.JWTAccountIDKey).(string)
	if !ok {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: Owner ID not found", "", nil, false)
		return
	}
	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid Owner ID format", "", nil, false)
		return
	}

	// Authorize ownership of the URL
	if err := authorizeURLOwnership(status.URLID, ownerID); err != nil {
		response.SendResponse(w, http.StatusUnauthorized, err.Error(), "", nil, false)
		return
	}

	// Validate percentage distribution
	if err := validatePercentageDistribution(status.URLID, status.Percentage); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Percentage validation failed", err.Error(), nil, false)
		return
	}

	// Update the HTTP status in the database
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
