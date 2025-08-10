package services

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

var db *sql.DB

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

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