package utils

import (
	"errors"
	"strings"
)

// Validatable allows structs to implement self-validation
type Validatable interface {
	Validate() error
}

// ValidateRequest calls Validate() if struct implements it
func ValidateRequest(v interface{}) error {
	if validatable, ok := v.(Validatable); ok {
		return validatable.Validate()
	}
	return nil
}

// Helper: required string
func RequiredString(value string, field string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(field + " is required")
	}
	return nil
}

// Helper: required positive number
func RequiredPositiveInt(value int64, field string) error {
	if value <= 0 {
		return errors.New(field + " must be positive")
	}
	return nil
}
