package router

import (
	"github.com/adolfooes/api_faker/internal/api/handler"
	"github.com/adolfooes/api_faker/internal/api/middleware"
	"github.com/gorilla/mux"
)

// InitializeRouter initializes the API routes
func InitializeRouter() *mux.Router {
	router := mux.NewRouter()

	// Public route (login)
	router.HandleFunc("/login", handler.LoginHandler).Methods("POST")
	router.HandleFunc("/account", handler.CreateAccountHandler).Methods("POST")

	// Protected routes
	securedRoutes := router.PathPrefix("/api").Subrouter()
	securedRoutes.Use(middleware.JWTMiddleware) // Apply JWT middleware

	// Account-related routes under /api
	securedRoutes.HandleFunc("/account", handler.GetAllAccountsHandler).Methods("GET")
	securedRoutes.HandleFunc("/account/{id:[0-9]+}", handler.GetAccountHandler).Methods("GET")
	securedRoutes.HandleFunc("/account/{id:[0-9]+}", handler.UpdateAccountHandler).Methods("PUT")
	securedRoutes.HandleFunc("/account/{id:[0-9]+}", handler.DeleteAccountHandler).Methods("DELETE")

	// Project-related routes under /api
	securedRoutes.HandleFunc("/project", handler.GetAllProjectsHandler).Methods("GET")
	securedRoutes.HandleFunc("/project/{id:[0-9]+}", handler.GetProjectHandler).Methods("GET")
	securedRoutes.HandleFunc("/project", handler.CreateProjectHandler).Methods("POST")
	securedRoutes.HandleFunc("/project/{id:[0-9]+}", handler.UpdateProjectHandler).Methods("PUT")
	securedRoutes.HandleFunc("/project/{id:[0-9]+}", handler.DeleteProjectHandler).Methods("DELETE")

	// URL Config-related routes under /api
	securedRoutes.HandleFunc("/url_config", handler.GetAllURLConfigsHandler).Methods("GET")
	securedRoutes.HandleFunc("/url_config/{id:[0-9]+}", handler.GetURLConfigHandler).Methods("GET")
	securedRoutes.HandleFunc("/url_config", handler.CreateURLConfigHandler).Methods("POST")
	securedRoutes.HandleFunc("/url_config/{id:[0-9]+}", handler.UpdateURLConfigHandler).Methods("PUT")
	securedRoutes.HandleFunc("/url_config/{id:[0-9]+}", handler.DeleteURLConfigHandler).Methods("DELETE")

	// URL HTTP Status-related routes under /api
	securedRoutes.HandleFunc("/url_http_status", handler.GetAllURLHTTPStatusesHandler).Methods("GET")
	securedRoutes.HandleFunc("/url_http_status", handler.CreateURLHTTPStatusHandler).Methods("POST")
	securedRoutes.HandleFunc("/url_http_status/{id:[0-9]+}", handler.UpdateURLHTTPStatusHandler).Methods("PUT")
	securedRoutes.HandleFunc("/url_http_status/{id:[0-9]+}", handler.DeleteURLHTTPStatusHandler).Methods("DELETE")

	// Response Model-related routes under /api
	securedRoutes.HandleFunc("/response_model", handler.GetAllResponseModelsHandler).Methods("GET")
	securedRoutes.HandleFunc("/response_model", handler.CreateResponseModelHandler).Methods("POST")
	securedRoutes.HandleFunc("/response_model/{id:[0-9]+}", handler.UpdateResponseModelHandler).Methods("PUT")
	securedRoutes.HandleFunc("/response_model/{id:[0-9]+}", handler.DeleteResponseModelHandler).Methods("DELETE")

	// Mock response route under /api
	securedRoutes.HandleFunc("/mock/{path:.*}", handler.MockHandler).Methods("GET", "POST", "PUT", "DELETE", "PATCH")

	return router
}
