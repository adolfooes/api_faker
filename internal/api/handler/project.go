package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/adolfooes/api_faker/config"
	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
)

// Project represents the project structure
type Project struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     int64  `json:"owner_id"` // Changed AccountID to OwnerID
}

func validateRequiredProjectFields(project Project) error {
	if project.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if len(project.Name) < 2 {
		return fmt.Errorf("project name must be at least 2 characters long")
	}
	return nil
}

func validateProjectOwnership(projectID int, ownerID int64) (bool, error) {
	filters := map[string]interface{}{
		"id":       projectID,
		"owner_id": ownerID,
	}
	projects, err := crud.List("project", filters)
	if err != nil {
		return false, err
	}
	if len(projects) == 0 {
		return false, nil
	}
	return true, nil
}

func validateDescription(description string) error {
	maxLength := 1000
	if len(description) > maxLength {
		return fmt.Errorf("description cannot be longer than %d characters", maxLength)
	}
	return nil
}

func sanitizeProjectInput(project *Project) {
	project.Name = strings.TrimSpace(project.Name)
	project.Description = strings.TrimSpace(project.Description)
}

// CreateProjectHandler handles the creation of a new project
func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var project Project

	// Decode the request body into the project struct
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	sanitizeProjectInput(&project)

	// Validate required fields
	if err := validateRequiredProjectFields(project); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Required fields validation failed", err.Error(), nil, false)
		return
	}

	// Validate required fields
	if err := validateDescription(project.Description); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Description validation failed", err.Error(), nil, false)
		return
	}

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

	// Add the owner ID to the project (for consistency in future use)
	project.OwnerID = ownerID

	// Insert the new project into the database, including the owner ID
	columns := []string{"name", "description", "owner_id"} // Updated to use owner_id
	values := []interface{}{project.Name, project.Description, ownerID}
	createdProject, err := crud.Create("project", columns, values) // Fetching the created project object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to create project", err.Error(), nil, false)
		return
	}

	// Send the created project in the response
	response.SendResponse(w, http.StatusCreated, "Project created successfully", "", createdProject, false)
}

// GetAllProjectsHandler retrieves all projects from the database
func GetAllProjectsHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all projects
	results, err := crud.List("project", filters)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve projects", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "Projects retrieved successfully", "", results, false)
}

// GetProjectHandler retrieves a single project by ID from the database
func GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL using Mux Vars
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", "ID is missing", nil, false)
		return
	}

	// Convert the ID string to an integer
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	// Query the database for the project by ID
	result, err := crud.Read("project", id)
	if err != nil {
		response.SendResponse(w, http.StatusNotFound, "Failed to retrieve project", err.Error(), nil, false)
		return
	}

	// Return the retrieved project
	response.SendResponse(w, http.StatusOK, "Project retrieved successfully", "", result, false)
}

func UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var project Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	// Validate project ID from URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.SendResponse(w, http.StatusBadRequest, "Invalid project ID", err.Error(), nil, false)
		return
	}

	// Query the project by ID
	_, err = crud.Read("project", id)
	if err != nil {
		response.SendResponse(w, http.StatusNotFound, "Project not found", err.Error(), nil, false)
		return
	}

	// Extract the owner ID from the context (injected by the JWT middleware)
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

	// Validate that the project belongs to the owner
	if err := authorizeProjectOwnership(project.ID, ownerID); err != nil {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: "+err.Error(), "", nil, false)
		return
	}

	// Update the project in the database using the crud package
	updates := map[string]interface{}{
		"name":        project.Name,
		"description": project.Description,
	}
	updatedProject, err := crud.Update("project", project.ID, updates) // Fetching the updated project object
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to update project", err.Error(), nil, false)
		return
	}

	// Send the updated project in the response
	response.SendResponse(w, http.StatusOK, "Project updated successfully", "", updatedProject, false)
}

func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	// Query the project by ID
	_, err = crud.Read("project", id)
	if err != nil {
		response.SendResponse(w, http.StatusNotFound, "Project not found", err.Error(), nil, false)
		return
	}

	// Extract the owner ID from the context (injected by the JWT middleware)
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

	// Validate that the project belongs to the owner
	if err := authorizeProjectOwnership(id, ownerID); err != nil {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: "+err.Error(), "", nil, false)
		return
	}

	// Perform the deletion
	err = crud.Delete("project", id)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to delete project", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "Project deleted successfully", "", nil, false)
}

// authorizeProjectOwnership validates if the current user owns the project.
func authorizeProjectOwnership(projectID int64, ownerID int64) error {
	filters := map[string]interface{}{
		"id":       projectID,
		"owner_id": ownerID,
	}

	projects, err := crud.List("project", filters)
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		return errors.New("you are not authorized to perform this operation on the project")
	}

	return nil
}
