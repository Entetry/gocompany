// Package handlers Contains rest handlers
package handlers

import "github.com/google/uuid"

type addCompanyRequest struct {
	Name string `json:"name" validate:"required"`
}

type updateCompanyRequest struct {
	UUID uuid.UUID `json:"uuid" validate:"required"`
	Name string    `json:"name" validate:"required"`
}
