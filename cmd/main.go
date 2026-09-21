package main

import (
	"log"
	"net/http"

	"github.com/adolfooes/api_faker/config"              // Import config package for DB connection strings
	"github.com/adolfooes/api_faker/internal/api/router" // Import the router
	"github.com/adolfooes/api_faker/internal/db"         // Import the database package if needed
)

func main() {
	db.InitDB(config.GetDatabaseConnectionString())
	db.RunMigrations(config.GetDatabaseConnectionString())

	jwtKey := config.GetJWTSecretKey()
	if len(jwtKey) == 0 || string(jwtKey) == "your_secret_key" {
		log.Println("WARNING: JWT_SECRET_KEY is not set or uses the insecure default value. Set a secure JWT_SECRET_KEY environment variable.")
	}

	router := router.InitializeRouter()

	log.Println("Server is running on port 8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
