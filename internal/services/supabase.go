package services

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

// db is a package-level variable for the database connection pool
var db *sql.DB

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// InitSupabase initializes the database connection
func InitSupabase() {
	var err error
	connStr := os.Getenv("SUPABASE_CONN_STRING")
	if connStr == "" {
		log.Fatal("SUPABASE_CONN_STRING not set")
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to Supabase: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging Supabase: %v", err)
	}
	log.Println("Successfully connected to Supabase!")
}