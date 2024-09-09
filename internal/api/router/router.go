package router

import (
	"github.com/adolfooes/api_faker/internal/api/handler"
	"github.com/gorilla/mux"
)

// InitializeRouter initializes the API routes
func InitializeRouter() *mux.Router {
	router := mux.NewRouter()

	// Account-related routes
	router.HandleFunc("/account", handler.GetAllAccountsHandler).Methods("GET")
	router.HandleFunc("/account/{id:[0-9]+}", handler.GetAccountHandler).Methods("GET")
	router.HandleFunc("/account", handler.CreateAccountHandler).Methods("POST")
	router.HandleFunc("/account/{id:[0-9]+}", handler.UpdateAccountHandler).Methods("PUT")
	router.HandleFunc("/account/{id:[0-9]+}", handler.DeleteAccountHandler).Methods("DELETE")

	// Project-related routes
	router.HandleFunc("/project", handler.GetAllProjectsHandler).Methods("GET")
	router.HandleFunc("/project/{id:[0-9]+}", handler.GetProjectHandler).Methods("GET")
	router.HandleFunc("/project", handler.CreateProjectHandler).Methods("POST")
	router.HandleFunc("/project/{id:[0-9]+}", handler.UpdateProjectHandler).Methods("PUT")
	router.HandleFunc("/project/{id:[0-9]+}", handler.DeleteProjectHandler).Methods("DELETE")

	// URL Config-related routes
	router.HandleFunc("/url_config", handler.GetAllURLConfigsHandler).Methods("GET")
	router.HandleFunc("/url_config/{id:[0-9]+}", handler.GetURLConfigHandler).Methods("GET")
	router.HandleFunc("/url_config", handler.CreateURLConfigHandler).Methods("POST")
	router.HandleFunc("/url_config/{id:[0-9]+}", handler.UpdateURLConfigHandler).Methods("PUT")
	router.HandleFunc("/url_config/{id:[0-9]+}", handler.DeleteURLConfigHandler).Methods("DELETE")

	// URL HTTP Status-related routes
	router.HandleFunc("/url_http_status", handler.GetAllURLHTTPStatusesHandler).Methods("GET")
	router.HandleFunc("/url_http_status", handler.CreateURLHTTPStatusHandler).Methods("POST")
	router.HandleFunc("/url_http_status/{id:[0-9]+}", handler.UpdateURLHTTPStatusHandler).Methods("PUT")
	router.HandleFunc("/url_http_status/{id:[0-9]+}", handler.DeleteURLHTTPStatusHandler).Methods("DELETE")

	// Response Model-related routes
	router.HandleFunc("/response_model", handler.GetAllResponseModelsHandler).Methods("GET")
	router.HandleFunc("/response_model", handler.CreateResponseModelHandler).Methods("POST")
	router.HandleFunc("/response_model/{id:[0-9]+}", handler.UpdateResponseModelHandler).Methods("PUT")
	router.HandleFunc("/response_model/{id:[0-9]+}", handler.DeleteResponseModelHandler).Methods("DELETE")

	return router
}
