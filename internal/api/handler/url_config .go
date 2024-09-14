package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

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

	// Insert the new URL config into the database using the crud package
	columns := []string{"path", "method", "description"}
	values := []interface{}{urlConfig.Path, urlConfig.Method, urlConfig.Description}
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

// UpdateURLConfigHandler handles updating an existing URL config
func UpdateURLConfigHandler(w http.ResponseWriter, r *http.Request) {
	var urlConfig URLConfig
	err := json.NewDecoder(r.Body).Decode(&urlConfig)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Update the URL config in the database using the crud package
	updates := map[string]interface{}{
		"path":        urlConfig.Path,
		"method":      urlConfig.Method,
		"description": urlConfig.Description,
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
