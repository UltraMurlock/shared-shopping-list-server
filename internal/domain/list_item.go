package domain

import (
	"time"
	"uuid"

	"github.com/shopspring/decimal"
)

type ListItem struct {
	ID        uuid.UUID
	ListID    uuid.UUID
	Name      string
	Comment   string
	UnitID    int16
	Quantity  decimal.Decimal
	IsDone    bool
	Version   int64
	CreatedBy uuid.UUID
	CreatedAt time.Time
	UpdateBy  uuid.UUID
	UpdateAt  time.Time
	DeletedBy *uuid.UUID
	DeletedAt *time.Time
}
