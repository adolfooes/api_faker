package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/adolfooes/api_faker/config"
	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
)

// URLConfig represents a URL configuration structure
type URLConfig struct {
	ID          int    `json:"id"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
	ProjectID   int    `json:"project_id"` // Add ProjectID to the struct
}

func validateRequiredURLConfigFields(urlConfig URLConfig) error {
	if urlConfig.Path == "" {
		return fmt.Errorf("path is required")
	}
	if urlConfig.Method == "" {
		return fmt.Errorf("method is required")
	}
	if urlConfig.ProjectID == 0 {
		return fmt.Errorf("project_id is required")
	}
	return nil
}

func validatePathFormat(path string) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with '/'")
	}
	// Optionally, add more regex checks for valid characters
	return nil
}

func validateHTTPMethod(method string) error {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return nil
		}
	}
	return fmt.Errorf("invalid HTTP method: %s", method)
}

func checkDuplicateURLConfig(projectID int, path string, method string) (bool, error) {
	filters := map[string]interface{}{
		"project_id": projectID,
		"path":       path,
		"method":     method,
	}
	urlConfigs, err := crud.List("url_config", filters)
	if err != nil {
		return false, err
	}
	if len(urlConfigs) > 0 {
		return true, nil
	}
	return false, nil
}

// CreateURLConfigHandler handles the creation of a new URL config
func CreateURLConfigHandler(w http.ResponseWriter, r *http.Request) {
	var urlConfig URLConfig

	// Decode the request body into the urlConfig struct
	err := json.NewDecoder(r.Body).Decode(&urlConfig)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Validate required fields
	if err := validateRequiredURLConfigFields(urlConfig); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Validation failed", err.Error(), nil, false)
		return
	}

	// Validate path format
	if err := validatePathFormat(urlConfig.Path); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid path format", err.Error(), nil, false)
		return
	}

	// Validate HTTP method
	if err := validateHTTPMethod(urlConfig.Method); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid HTTP method", err.Error(), nil, false)
		return
	}

	// Extract the owner ID from the context (JWT middleware)
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

	// Authorize the project ownership
	if err := authorizeProjectOwnership(urlConfig.ProjectID, ownerID); err != nil {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: "+err.Error(), "", nil, false)
		return
	}

	// Check for duplicate URL config (same path and method within the same project)
	exists, err := checkDuplicateURLConfig(urlConfig.ProjectID, urlConfig.Path, urlConfig.Method)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Error checking duplicate URL config", err.Error(), nil, false)
		return
	}
	if exists {
		response.SendResponse(w, http.StatusConflict, "URL config with the same path and method already exists", "", nil, false)
		return
	}

	// Insert the new URL config into the database
	columns := []string{"path", "method", "description", "project_id"}
	values := []interface{}{urlConfig.Path, urlConfig.Method, urlConfig.Description, urlConfig.ProjectID}
	createdConfig, err := crud.Create("url_config", columns, values) // Fetch the created object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to create URL config", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusCreated, "URL config created successfully", "", createdConfig, false)
}

// GetAllURLConfigsHandler retrieves all URL configs from the database
func GetAllURLConfigsHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all URL configs
	results, err := crud.List("url_config", filters)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve URL configs", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "URL configs retrieved successfully", "", results, false)
}

// GetURLConfigHandler retrieves a single URL config by ID from the database
func GetURLConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL path using mux.Vars
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", "ID is missing", nil, false)
		return
	}

	// Convert the ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	// Read the URL config from the database
	result, err := crud.Read("url_config", id)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve URL config", err.Error(), nil, false)
		return
	}

	// Send a successful response
	response.SendResponse(w, http.StatusOK, "URL config retrieved successfully", "", result, false)
}

func UpdateURLConfigHandler(w http.ResponseWriter, r *http.Request) {
	var urlConfig URLConfig
	err := json.NewDecoder(r.Body).Decode(&urlConfig)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Validate required fields
	if err := validateRequiredURLConfigFields(urlConfig); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Validation failed", err.Error(), nil, false)
		return
	}

	// Validate path format
	if err := validatePathFormat(urlConfig.Path); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid path format", err.Error(), nil, false)
		return
	}

	// Validate HTTP method
	if err := validateHTTPMethod(urlConfig.Method); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid HTTP method", err.Error(), nil, false)
		return
	}

	// Extract the owner ID from the context (JWT middleware)
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

	// Authorize the project ownership
	if err := authorizeProjectOwnership(urlConfig.ProjectID, ownerID); err != nil {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: "+err.Error(), "", nil, false)
		return
	}

	// Update the URL config in the database
	updates := map[string]interface{}{
		"path":        urlConfig.Path,
		"method":      urlConfig.Method,
		"description": urlConfig.Description,
		"project_id":  urlConfig.ProjectID,
	}
	updatedConfig, err := crud.Update("url_config", urlConfig.ID, updates) // Fetch the updated object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to update URL config", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "URL config updated successfully", "", updatedConfig, false)
}

// DeleteURLConfigHandler handles deleting a URL config
func DeleteURLConfigHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	err = crud.Delete("url_config", id)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to delete URL config", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "URL config deleted successfully", "", nil, false)
}
