package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/adolfooes/api_faker/config"
	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
)

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	// Regular expression to validate the path
	// This regex allows alphanumeric characters, slashes (/), dashes (-), and underscores (_)
	regex := regexp.MustCompile(`^\/[a-zA-Z0-9\/\-_]*$`)

	if !regex.MatchString(path) {
		return fmt.Errorf("invalid path: path can only contain alphanumeric characters, slashes (/), dashes (-), and underscores (_)")
	}

	return nil
}

func validateOwnerID(ownerID string) (int64, error) {
	if ownerID == "" {
		return 0, fmt.Errorf("owner ID is required")
	}

	ownerIDInt, err := strconv.ParseInt(ownerID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid Owner ID format")
	}

	return ownerIDInt, nil
}

func validateProjectID(projectIDStr string) (int, error) {
	if projectIDStr == "" {
		return 0, fmt.Errorf("missing project ID in URL")
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid project ID")
	}

	return projectID, nil
}

func checkProjectOwnership(project map[string]interface{}, ownerID int64) error {
	if project["owner_id"].(int64) != ownerID {
		return fmt.Errorf("unauthorized: owner ID mismatch")
	}
	return nil
}

func fetchProject(projectID int) (map[string]interface{}, error) {
	filters := map[string]interface{}{
		"id": projectID,
	}
	projectResults, err := crud.List("project", filters)
	if err != nil || len(projectResults) == 0 {
		return nil, fmt.Errorf("project not found")
	}

	return projectResults[0], nil
}

func MockHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the requested path from the URL
	vars := mux.Vars(r)
	path := "/" + vars["path"]

	// Validate the path
	if err := validatePath(path); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid path", err.Error(), nil, false)
		return
	}

	// Extract the owner ID (which will be used as owner_id) from the context (injected by the JWT middleware)
	ownerIDStr, ok := r.Context().Value(config.JWTAccountIDKey).(string)
	if !ok {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: Owner ID not found", "", nil, false)
		return
	}

	// Validate the owner ID
	ownerID, err := validateOwnerID(ownerIDStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid Owner ID format", err.Error(), nil, false)
		return
	}

	// Extract the project ID from the URL parameters
	projectIDStr := vars["project_id"]

	// Validate project ID
	projectID, err := validateProjectID(projectIDStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid project ID", err.Error(), nil, false)
		return
	}

	// Fetch the project from the database
	project, err := fetchProject(projectID)
	if err != nil {
		response.SendResponse(w, http.StatusNotFound, "Project not found", err.Error(), nil, false)
		return
	}

	// Check ownership of the project
	if err := checkProjectOwnership(project, ownerID); err != nil {
		response.SendResponse(w, http.StatusUnauthorized, err.Error(), "", nil, false)
		return
	}

	// Check if the URL is configured in the database for the given project
	method := strings.ToUpper(r.Method)
	filters := map[string]interface{}{
		"path":       path,
		"method":     method,
		"project_id": projectID, // Filter by the project ID from URL
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
