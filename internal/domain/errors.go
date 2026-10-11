package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrVersionConflict    = errors.New("version conflict")
	ErrValidationError    = errors.New("validation error")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
