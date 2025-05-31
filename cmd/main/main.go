package main

import (
	"fmt"
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

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	fmt.Println("-------------------Start--------------")
	config.ConnectDB()
	fmt.Println("-------------------Close--------------")

	defer config.DB.Close()

	router := routes.SetupRouter()
	log.Fatal(router.Run(":8080"))
}
