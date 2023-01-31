// Package middleware contains middleware filter structures
package middleware

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator validation middleware
type CustomValidator struct {
	Validator *validator.Validate
}

// NewCustomValidator creates CustomValidator object
func NewCustomValidator(v *validator.Validate) *CustomValidator {
	return &CustomValidator{
		Validator: v,
	}
}

// Validate validates any object by go-playground/validator/v10 tags
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
