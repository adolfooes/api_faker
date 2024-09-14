package handler

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/adolfooes/api_faker/config"
	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
)

// MockHandler handles incoming requests for mocked URLs and returns the appropriate mocked response
func MockHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the requested path from the URL
	vars := mux.Vars(r)
	path := "/" + vars["path"]

	// Extract the account ID (which will be used as owner_id) from the context (injected by the JWT middleware)
	ownerIDStr, ok := r.Context().Value(config.JWTAccountIDKey).(string)
	if !ok {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: Owner ID not found", "", nil, false)
		return
	}

	// Convert the ownerID from string to int64
	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid Owner ID format", "", nil, false)
		return
	}

	// Get the project_id from the headers
	projectIDStr := r.Header.Get("Project-ID")
	if projectIDStr == "" {
		response.SendResponse(w, http.StatusBadRequest, "Missing project ID in headers", "", nil, false)
		return
	}

	// Convert the project ID to an integer
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid project ID", "", nil, false)
		return
	}

	// Fetch the project from the database
	projectFilters := map[string]interface{}{
		"id": projectID,
	}
	projectResults, err := crud.List("project", projectFilters)
	if err != nil || len(projectResults) == 0 {
		response.SendResponse(w, http.StatusNotFound, "Project not found", err.Error(), nil, false)
		return
	}
	project := projectResults[0]

	// Compare the owner_id from the project with the one from the JWT claims
	if project["owner_id"].(int64) != ownerID {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: Owner ID mismatch", "", nil, false)
		return
	}

	// Check if the URL is configured in the database for the given project
	method := strings.ToUpper(r.Method)
	filters := map[string]interface{}{
		"path":       path,
		"method":     method,
		"project_id": projectID, // Filter by the project ID from headers
	}
	urlConfigs, err := crud.List("url_config", filters)
	if err != nil || len(urlConfigs) == 0 {
		response.SendResponse(w, http.StatusNotFound, "URL not configured for mocking", "", nil, false)
		return
	}
	urlConfig := urlConfigs[0]

	// Fetch all the HTTP statuses and their percentages from url_http_status for this url_config
	filters = map[string]interface{}{
		"url_id": urlConfig["id"],
	}
	httpStatuses, err := crud.List("url_http_status", filters)
	if err != nil || len(httpStatuses) == 0 {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to fetch HTTP statuses", err.Error(), nil, false)
		return
	}

	// Randomize the response based on percentage
	selectedStatus := randomizeHTTPStatus(httpStatuses)

	// Fetch the corresponding response model from the response_model table
	filters = map[string]interface{}{
		"url_http_status_id": selectedStatus["id"],
	}
	responseModels, err := crud.List("response_model", filters)
	if err != nil || len(responseModels) == 0 {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to fetch response model", err.Error(), nil, false)
		return
	}
	responseModel := responseModels[0]

	// Send the mock response using the SendResponse function
	response.SendResponse(w, int(selectedStatus["http_status"].(int64)), "", "", responseModel["model"], true)
}

// randomizeHTTPStatus selects a status based on the percentage distribution
func randomizeHTTPStatus(statuses []map[string]interface{}) map[string]interface{} {
	totalPercentage := 0
	for _, status := range statuses {
		totalPercentage += int(status["percentage"].(int64))
	}

	randomNumber := rand.Intn(100) // Random number between 0 and 99
	currentPercentage := 0

	for _, status := range statuses {
		currentPercentage += int(status["percentage"].(int64))
		if randomNumber < currentPercentage {
			return status
		}
	}

	// Default to the first one if something goes wrong
	return statuses[0]
}
