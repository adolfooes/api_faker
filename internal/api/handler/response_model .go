package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
)

// ResponseModel represents a structure for a response model
type ResponseModel struct {
	ID              int         `json:"id"`
	URLHTTPStatusID int         `json:"url_http_status_id"`
	Model           interface{} `json:"model"` // Assuming model is JSONB in the database
	Description     string      `json:"description"`
}

// CreateResponseModelHandler handles creating a new response model for an HTTP status
func CreateResponseModelHandler(w http.ResponseWriter, r *http.Request) {
	var model ResponseModel

	// Decode the request body into the response model struct
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Insert the new response model into the database using the crud package
	columns := []string{"url_http_status_id", "model", "description"}
	values := []interface{}{model.URLHTTPStatusID, model.Model, model.Description}
	createdModel, err := crud.Create("response_model", columns, values) // Fetch the created response model object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to create response model", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusCreated, "Response model created successfully", "", createdModel, false)
}

// GetAllResponseModelsHandler retrieves all response models from the database
func GetAllResponseModelsHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all response models
	results, err := crud.List("response_model", filters)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve response models", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "Response models retrieved successfully", "", results, false)
}

// GetResponseModelHandler retrieves a single response model by ID from the database
func GetResponseModelHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL using Mux Vars
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", "ID is missing", nil, false)
		return
	}

	// Convert the ID string to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	// Query the database for the response model by ID
	result, err := crud.Read("response_model", id)
	if err != nil {
		response.SendResponse(w, http.StatusNotFound, "Failed to retrieve response model", err.Error(), nil, false)
		return
	}

	// Return the retrieved response model
	response.SendResponse(w, http.StatusOK, "Response model retrieved successfully", "", result, false)
}

// UpdateResponseModelHandler handles updating an existing response model
func UpdateResponseModelHandler(w http.ResponseWriter, r *http.Request) {
	var model ResponseModel
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Update the response model in the database using the crud package
	updates := map[string]interface{}{
		"url_http_status_id": model.URLHTTPStatusID,
		"model":              model.Model,
		"description":        model.Description,
	}
	updatedModel, err := crud.Update("response_model", model.ID, updates) // Fetch the updated response model object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to update response model", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "Response model updated successfully", "", updatedModel, false)
}

// DeleteResponseModelHandler handles deleting a response model
func DeleteResponseModelHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	err = crud.Delete("response_model", id)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to delete response model", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "Response model deleted successfully", "", nil, false)
}
