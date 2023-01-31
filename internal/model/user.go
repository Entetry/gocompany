package model

import "github.com/google/uuid"

// User user domain model
type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
}
