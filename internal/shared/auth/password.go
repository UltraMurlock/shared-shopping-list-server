package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordMismatch = errors.New("password mismatch")

const bcryptCost = 12

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}

	return string(hash), nil
}

func CheckPassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrPasswordMismatch
	}

	if err != nil {
		return fmt.Errorf("auth: chech password: %w", err)
	}

	return nil
}
