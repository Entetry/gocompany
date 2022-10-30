// Package model domain models package
package model

import "github.com/google/uuid"

// Company domain company struct
type Company struct {
	ID   uuid.UUID `bson:"_id"`
	Name string    `bson:"name"`
}
