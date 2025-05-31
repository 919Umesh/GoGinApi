package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/umesh/ginapi/config"
	"github.com/umesh/ginapi/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Validate JWT secret
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Connect to database
	config.ConnectDB()
	defer config.DB.Close()

	// Setup and run router
	router := routes.SetupRouter()
	log.Fatal(router.Run(":8080"))
}
