package model

import "github.com/google/uuid"

// Logo company logo domain model
type Logo struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Image     string
}
