package main

import (
	"log"
	"net/http"

	"github.com/adolfooes/api_faker/config"              // Import config package for DB connection strings
	"github.com/adolfooes/api_faker/internal/api/router" // Import the router
	"github.com/adolfooes/api_faker/internal/db"         // Import the database package if needed
)

func main() {
	// Initialize the database connection (if you're using a database)
	db.InitDB(config.GetDatabaseConnectionString())

	// Initialize the router
	router := router.InitializeRouter()

	// Start the HTTP server
	log.Println("Server is running on port 8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
