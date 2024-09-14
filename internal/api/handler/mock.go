package handler

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
)

// MockHandler handles incoming requests for mocked URLs and returns the appropriate mocked response
func MockHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the requested path from the URL
	vars := mux.Vars(r)
	path := "/" + vars["path"]

	// Get the project_id from the JWT claims (assuming middleware injects the account info into context)
	ownerID, ok := r.Context().Value("owner_id").(string) // Replace with your context key
	if !ok {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: Owner ID not found", "", nil, false)
		return
	}

	// Check if the URL is configured in the database for the given project
	method := strings.ToUpper(r.Method)
	filters := map[string]interface{}{
		"path":       path,
		"method":     method,
		"project_id": ownerID, // Filter by the project ID
	}
	urlConfigs, err := crud.List("url_config", filters)
	if err != nil || len(urlConfigs) == 0 {
		response.SendResponse(w, http.StatusNotFound, "URL not configured for mocking", "", nil, false)
		return
	}
	urlConfig := urlConfigs[0]

	// Fetch all the http statuses and their percentage from url_http_status for this url_config
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

	// Return the mocked response
	w.WriteHeader(int(selectedStatus["http_status"].(int64)))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseModel["model"])
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
