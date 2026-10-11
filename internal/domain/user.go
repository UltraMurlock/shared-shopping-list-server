package domain

import (
	"time"
	"uuid"
)

type User struct {
	ID           uuid.UUID
	Login        string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
	UpdateAt     time.Time
}
