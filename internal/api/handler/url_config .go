package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
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
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Insert the new URL config into the database using the crud package
	columns := []string{"path", "method", "description"}
	values := []interface{}{urlConfig.Path, urlConfig.Method, urlConfig.Description}
	err = crud.Create("url_config", columns, values)
	if err != nil {
		response.SendResponse(w, "Failed to create URL config", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "URL config created successfully", "", urlConfig, nil)
}

// GetAllURLConfigsHandler retrieves all URL configs from the database
func GetAllURLConfigsHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all URL configs
	results, err := crud.List("url_config", filters)
	if err != nil {
		response.SendResponse(w, "Failed to retrieve URL configs", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "URL configs retrieved successfully", "", results, nil)
}

// GetURLConfigHandler retrieves a single URL config by ID from the database
func GetURLConfigHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, "Invalid ID parameter", err.Error(), nil, nil)
		return
	}

	result, err := crud.Read("url_config", id)
	if err != nil {
		response.SendResponse(w, "Failed to retrieve URL config", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "URL config retrieved successfully", "", result, nil)
}

// UpdateURLConfigHandler handles updating an existing URL config
func UpdateURLConfigHandler(w http.ResponseWriter, r *http.Request) {
	var urlConfig URLConfig
	err := json.NewDecoder(r.Body).Decode(&urlConfig)
	if err != nil {
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Update the URL config in the database using the crud package
	updates := map[string]interface{}{
		"path":        urlConfig.Path,
		"method":      urlConfig.Method,
		"description": urlConfig.Description,
	}
	err = crud.Update("url_config", urlConfig.ID, updates)
	if err != nil {
		response.SendResponse(w, "Failed to update URL config", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "URL config updated successfully", "", urlConfig, nil)
}

// DeleteURLConfigHandler handles deleting a URL config
func DeleteURLConfigHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, "Invalid ID parameter", err.Error(), nil, nil)
		return
	}

	err = crud.Delete("url_config", id)
	if err != nil {
		response.SendResponse(w, "Failed to delete URL config", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "URL config deleted successfully", "", nil, nil)
}
