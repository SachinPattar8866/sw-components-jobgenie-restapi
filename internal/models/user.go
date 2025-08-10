package models

import "time"

type User struct {
    ID          string    `json:"id"`
    FirebaseUID string    `json:"firebase_uid"`
    Email       string    `json:"email"`
    FullName    string    `json:"full_name"`
    CreatedAt   time.Time `json:"created_at"`
}