package services

import (
	"context"
	"fmt"
	"log"

	"sw-components-jobgenie-restapi/internal/models"
)

// CreateUser inserts a new user into the database
func CreateUser(ctx context.Context, user models.User) error {
	query := `INSERT INTO users (firebase_uid, email, full_name) VALUES ($1, $2, $3) ON CONFLICT (firebase_uid) DO NOTHING;`
	_, err := db.ExecContext(ctx, query, user.FirebaseUID, user.Email, user.FullName)
	if err != nil {
		log.Printf("Error creating user: %v", err)
	}
	return err
}

// UserExists checks if a user with the given Firebase UID already exists
func UserExists(ctx context.Context, firebaseUID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE firebase_uid = $1`
	err := db.QueryRowContext(ctx, query, firebaseUID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to query user existence: %w", err)
	}
	return count > 0, nil
}