package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/umesh/ginapi/config"
	"github.com/umesh/ginapi/routes"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("start")
	config.ConnectDB()
	fmt.Println("end")
	router := routes.SetupRouter()

	// Run server
	router.Run(":8080")
}
