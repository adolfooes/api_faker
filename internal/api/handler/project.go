package handler

import (
	"encoding/json"
	"net/http"
)

// GetProjects handles the request to retrieve all projects
func GetProjects(w http.ResponseWriter, r *http.Request) {
	projects := []map[string]interface{}{
		{"id": 1, "name": "Project A"},
		{"id": 2, "name": "Project B"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// GetProject handles the request to retrieve a single project by ID
func GetProject(w http.ResponseWriter, r *http.Request) {
	// Example of retrieving project by ID
	project := map[string]interface{}{"id": 1, "name": "Project A"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// CreateProject handles the creation of a new project
func CreateProject(w http.ResponseWriter, r *http.Request) {
	// Logic to create a new project
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Project created"))
}

// UpdateProject handles the updating of an existing project by ID
func UpdateProject(w http.ResponseWriter, r *http.Request) {
	// Logic to update a project by ID
	w.Write([]byte("Project updated"))
}

// DeleteProject handles the deletion of a project by ID
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	// Logic to delete a project by ID
	w.Write([]byte("Project deleted"))
}
