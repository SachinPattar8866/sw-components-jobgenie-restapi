package models

import (
	"time"

	"github.com/google/uuid"
)

type Resume struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	FileURL     string    `json:"file_url" db:"file_url"`
	TextContent string    `json:"text_content" db:"text_content"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
