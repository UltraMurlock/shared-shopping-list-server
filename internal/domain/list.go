package domain

import (
	"time"
	"uuid"
)

type List struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Description string
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
