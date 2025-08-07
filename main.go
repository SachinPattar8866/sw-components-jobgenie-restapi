package main

import (
	"log"
	"os"

	"sw-components-jobgenie-restapi/internal"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := internal.InitServer()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on port %s", port)
	router.Run(":" + port)
}
