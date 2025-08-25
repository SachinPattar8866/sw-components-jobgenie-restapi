package services

import (
	"context"
	"fmt"
	"log"

	"sw-components-jobgenie-restapi/internal/models"
)

// UserService manages user-related business logic.
type UserService struct {
	Supabase *Client // Now correctly references the Supabase client
}

// NewUserService creates a new instance of UserService.
func NewUserService(sc *Client) *UserService {
	return &UserService{
		Supabase: sc,
	}
}

// CreateUser inserts a new user into the database.
func (s *UserService) CreateUser(ctx context.Context, user models.User) error {
	query := `INSERT INTO users (firebase_uid, email, full_name) VALUES ($1, $2, $3) ON CONFLICT (firebase_uid) DO NOTHING;`
	_, err := s.Supabase.DB.ExecContext(ctx, query, user.FirebaseUID, user.Email, user.FullName)
	if err != nil {
		log.Printf("Error creating user: %v", err)
	}
	return err
}

// UserExists checks if a user with the given Firebase UID exists.
func (s *UserService) UserExists(ctx context.Context, firebaseUID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE firebase_uid = $1`
	err := s.Supabase.DB.QueryRowContext(ctx, query, firebaseUID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to query user existence: %w", err)
	}
	return count > 0, nil
}