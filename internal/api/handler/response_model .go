package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
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
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Insert the new response model into the database using the crud package
	columns := []string{"url_http_status_id", "model", "description"}
	values := []interface{}{model.URLHTTPStatusID, model.Model, model.Description}
	err = crud.Create("response_model", columns, values)
	if err != nil {
		response.SendResponse(w, "Failed to create response model", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Response model created successfully", "", model, nil)
}

// GetAllResponseModelsHandler retrieves all response models from the database
func GetAllResponseModelsHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all response models
	results, err := crud.List("response_model", filters)
	if err != nil {
		response.SendResponse(w, "Failed to retrieve response models", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Response models retrieved successfully", "", results, nil)
}

// UpdateResponseModelHandler handles updating an existing response model
func UpdateResponseModelHandler(w http.ResponseWriter, r *http.Request) {
	var model ResponseModel
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Update the response model in the database using the crud package
	updates := map[string]interface{}{
		"url_http_status_id": model.URLHTTPStatusID,
		"model":              model.Model,
		"description":        model.Description,
	}
	err = crud.Update("response_model", model.ID, updates)
	if err != nil {
		response.SendResponse(w, "Failed to update response model", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Response model updated successfully", "", model, nil)
}

// DeleteResponseModelHandler handles deleting a response model
func DeleteResponseModelHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, "Invalid ID parameter", err.Error(), nil, nil)
		return
	}

	err = crud.Delete("response_model", id)
	if err != nil {
		response.SendResponse(w, "Failed to delete response model", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Response model deleted successfully", "", nil, nil)
}
