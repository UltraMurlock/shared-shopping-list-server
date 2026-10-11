package domain

import (
	"encoding/json"
	"time"
	"uuid"
)

type EntityType string

const (
	EntityTypeList EntityType = "list"
	EntityTypeItem EntityType = "item"
)

type Operation string

const (
	OperationCreated Operation = "created"
	OperationUpdated Operation = "updated"
	OperationDeleted Operation = "deleted"
)

type ListEvent struct {
	ID            int64
	ListID        uuid.UUID
	ActorID       uuid.UUID
	EntityType    EntityType
	EntityID      uuid.UUID
	Operation     Operation
	EntityVersion int64
	Payload       json.RawMessage
	CreatedAt     time.Time
}
