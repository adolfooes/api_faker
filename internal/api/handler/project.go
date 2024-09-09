package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
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
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Insert the new project into the database using the crud package
	columns := []string{"name", "description"}
	values := []interface{}{project.Name, project.Description}
	err = crud.Create("project", columns, values)
	if err != nil {
		response.SendResponse(w, "Failed to create project", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Project created successfully", "", project, nil)
}

// GetAllProjectsHandler retrieves all projects from the database
func GetAllProjectsHandler(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{} // No filters, get all projects
	results, err := crud.List("project", filters)
	if err != nil {
		response.SendResponse(w, "Failed to retrieve projects", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Projects retrieved successfully", "", results, nil)
}

// GetProjectHandler retrieves a single project by ID from the database
func GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, "Invalid ID parameter", err.Error(), nil, nil)
		return
	}

	result, err := crud.Read("project", id)
	if err != nil {
		response.SendResponse(w, "Failed to retrieve project", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Project retrieved successfully", "", result, nil)
}

// UpdateProjectHandler handles updating an existing project
func UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var project Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		response.SendResponse(w, "Invalid request payload", err.Error(), nil, nil)
		return
	}

	// Update the project in the database using the crud package
	updates := map[string]interface{}{
		"name":        project.Name,
		"description": project.Description,
	}
	err = crud.Update("project", project.ID, updates)
	if err != nil {
		response.SendResponse(w, "Failed to update project", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Project updated successfully", "", project, nil)
}

// DeleteProjectHandler handles deleting a project
func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.SendResponse(w, "Invalid ID parameter", err.Error(), nil, nil)
		return
	}

	err = crud.Delete("project", id)
	if err != nil {
		response.SendResponse(w, "Failed to delete project", err.Error(), nil, nil)
		return
	}

	response.SendResponse(w, "Project deleted successfully", "", nil, nil)
}
