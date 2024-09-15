package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/config"
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

func validateRequiredResponseModelFields(model ResponseModel) error {
	if model.URLHTTPStatusID == 0 {
		return fmt.Errorf("url_http_status_id is required")
	}
	if model.Model == nil {
		return fmt.Errorf("model is required")
	}
	return nil
}

func validateURLHTTPStatusExists(urlHTTPStatusID int) error {
	filters := map[string]interface{}{"id": urlHTTPStatusID}
	statuses, err := crud.List("url_http_status", filters)
	if err != nil || len(statuses) == 0 {
		return fmt.Errorf("url_http_status_id does not exist")
	}
	return nil
}

func authorizeOwnershipByURLHTTPStatusID(urlHTTPStatusID int, ownerID int64) error {
	// First, fetch the url_id associated with the provided url_http_status_id
	filters := map[string]interface{}{"id": urlHTTPStatusID}
	urlHTTPStatuses, err := crud.List("url_http_status", filters)
	if err != nil || len(urlHTTPStatuses) == 0 {
		return fmt.Errorf("url_http_status_id not found")
	}
	urlHTTPStatus := urlHTTPStatuses[0]

	// Now, fetch the corresponding url_config using the url_id to check ownership
	urlID := urlHTTPStatus["url_id"].(int)
	filters = map[string]interface{}{"id": urlID}
	urlConfigs, err := crud.List("url_config", filters)
	if err != nil || len(urlConfigs) == 0 {
		return fmt.Errorf("url_config not found for the provided url_http_status_id")
	}

	// Check ownership of the url_config
	if urlConfigs[0]["owner_id"].(int64) != ownerID {
		return fmt.Errorf("you are not authorized to modify this response model")
	}

	return nil
}

func CreateResponseModelHandler(w http.ResponseWriter, r *http.Request) {
	var model ResponseModel

	// Decode the request body into the response model struct
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Validate required fields
	if err := validateRequiredResponseModelFields(model); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Validation failed", err.Error(), nil, false)
		return
	}

	// Validate the existence of url_http_status_id
	if err := validateURLHTTPStatusExists(model.URLHTTPStatusID); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid url_http_status_id", err.Error(), nil, false)
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

	// Authorize ownership by checking the owner of the url_http_status_id
	if err := authorizeOwnershipByURLHTTPStatusID(model.URLHTTPStatusID, ownerID); err != nil {
		response.SendResponse(w, http.StatusUnauthorized, err.Error(), "", nil, false)
		return
	}

	// Insert the new response model into the database
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

func UpdateResponseModelHandler(w http.ResponseWriter, r *http.Request) {
	var model ResponseModel
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Validate required fields
	if err := validateRequiredResponseModelFields(model); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Validation failed", err.Error(), nil, false)
		return
	}

	// Validate the existence of url_http_status_id
	if err := validateURLHTTPStatusExists(model.URLHTTPStatusID); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid url_http_status_id", err.Error(), nil, false)
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

	// Authorize ownership by checking the owner of the url_http_status_id
	if err := authorizeOwnershipByURLHTTPStatusID(model.URLHTTPStatusID, ownerID); err != nil {
		response.SendResponse(w, http.StatusUnauthorized, err.Error(), "", nil, false)
		return
	}

	// Update the response model in the database
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
