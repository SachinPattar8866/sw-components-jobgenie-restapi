package main

import (
	"log"
	"os"

	"sw-components-jobgenie-restapi/internal"
	"sw-components-jobgenie-restapi/internal/handlers"
	"sw-components-jobgenie-restapi/internal/services"

	"github.com/joho/godotenv"
	unipdflic "github.com/unidoc/unipdf/v3/common/license"
)

func main() {
	// Load env
	if err := godotenv.Load(".env"); err != nil {
		log.Println("warn: .env not found, relying on environment")
	}

	// Init Firebase
	services.InitFirebase()

	// Initialize UniPDF license (if provided)
	if key := os.Getenv("UNIDOC_LICENSE_KEY"); key != "" {
		if err := unipdflic.SetMeteredKey(key); err != nil {
			log.Printf("warning: failed to set UniPDF license: %v", err)
		} else {
			log.Println("UniPDF license set from UNIDOC_LICENSE_KEY")
		}
	} else {
		log.Println("UNIDOC_LICENSE_KEY not set; PDF parsing may require a license")
	}

	// Init Supabase (DB + Storage)
	sb, err := services.NewSupabaseClient()
	if err != nil {
		log.Fatalf("supabase init: %v", err)
	}
	defer sb.DB.Close()

	// Handlers
	userHandler := handlers.NewUserHandler(services.NewUserService(sb))
	resumeHandler := handlers.NewResumeHandler(services.NewResumeService(sb))

	// Server
	router := internal.InitServer(userHandler, resumeHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
