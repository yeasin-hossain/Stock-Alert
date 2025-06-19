package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/hello-api/internal/db"
	"github.com/hello-api/internal/router"
)

func main() {
	// Load environment variables
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev" // Default to development environment
		log.Println("ENV not set, defaulting to dev")
	}

	var envFile string
	switch env {
	case "dev":
		envFile = "config/env/dev.env"
	case "prod":
		envFile = "config/env/prod.env"
	default:
		log.Fatalf("Unknown ENV: %s", env)
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: Error loading env file: %v", err)
		log.Println("Continuing with default or existing environment variables")
	}

	// MongoDB URI is now hardcoded in the ConnectMongo function

	// Connect to MongoDB
	mongoClient := db.GetClient()
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Error disconnecting MongoDB: %v", err)
		}
	}()

	// Initialize routes
	r := router.InitializeRoutes()

	// Set up the server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Starting server on port 8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
