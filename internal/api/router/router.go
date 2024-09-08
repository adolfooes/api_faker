package router

import (
	"github.com/adolfooes/api_faker/internal/api/handler"
	"github.com/gorilla/mux"
)

// InitializeRouter initializes the API routes
func InitializeRouter() *mux.Router {
	router := mux.NewRouter()

	// Account-related routes
	router.HandleFunc("/accounts", handler.GetAccounts).Methods("GET")
	router.HandleFunc("/accounts/{id:[0-9]+}", handler.GetAccount).Methods("GET")
	router.HandleFunc("/accounts", handler.CreateAccount).Methods("POST")
	router.HandleFunc("/accounts/{id:[0-9]+}", handler.UpdateAccount).Methods("PUT")
	router.HandleFunc("/accounts/{id:[0-9]+}", handler.DeleteAccount).Methods("DELETE")

	// Project-related routes
	router.HandleFunc("/projects", handler.GetProjects).Methods("GET")
	router.HandleFunc("/projects/{id:[0-9]+}", handler.GetProject).Methods("GET")
	router.HandleFunc("/projects", handler.CreateProject).Methods("POST")
	router.HandleFunc("/projects/{id:[0-9]+}", handler.UpdateProject).Methods("PUT")
	router.HandleFunc("/projects/{id:[0-9]+}", handler.DeleteProject).Methods("DELETE")

	// Mock-related routes (for testing)
	router.HandleFunc("/mocks", handler.GetMock).Methods("GET")
	router.HandleFunc("/mocks", handler.CreateMock).Methods("POST")
	router.HandleFunc("/mocks", handler.UpdateMock).Methods("PUT")
	router.HandleFunc("/mocks", handler.DeleteMock).Methods("DELETE")

	return router
}
