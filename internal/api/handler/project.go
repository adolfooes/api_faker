package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
)

// Project represents the project structure
type Project struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
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

	// Insert the new project into the database using the crud package
	columns := []string{"name", "description"}
	values := []interface{}{project.Name, project.Description}
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
	id, err := strconv.Atoi(idStr)
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

// UpdateProjectHandler handles updating an existing project
func UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var project Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
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

// DeleteProjectHandler handles deleting a project
func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid ID parameter", err.Error(), nil, false)
		return
	}

	err = crud.Delete("project", id)
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to delete project", err.Error(), nil, false)
		return
	}

	response.SendResponse(w, http.StatusOK, "Project deleted successfully", "", nil, false)
}
