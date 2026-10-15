package handler

import (
	"encoding/json"
	"shared-shopping-list-server/internal/domain"
	"time"
	"uuid"
)

type eventResponse struct {
	ID            int64          `json:"id"`
	ListID        uuid.UUID      `json:"list_id"`
	ActorID       uuid.UUID      `json:"actor_id"`
	EntityType    string         `json:"entity_type"`
	EntityID      uuid.UUID      `json:"entity_id"`
	EntityVersion int64          `json:"entity_version"`
	Operation     string         `json:"operation"`
	Payload       map[string]any `json:"payload"`
	CreatedAt     time.Time      `json:"created_at"`
}

func toEventResponse(e domain.ListEvent) eventResponse {
	var payload map[string]any
	if len(e.Payload) > 0 {
		_ = json.Unmarshal(e.Payload, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}

	return eventResponse{
		ID:            e.ID,
		ListID:        e.ListID,
		ActorID:       e.ActorID,
		EntityType:    string(e.EntityType),
		EntityID:      e.EntityID,
		EntityVersion: e.EntityVersion,
		Operation:     string(e.Operation),
		Payload:       payload,
		CreatedAt:     e.CreatedAt,
	}
}

type listEventsResponse struct {
	Events  []eventResponse `json:"events"`
	HasMore bool            `json:"has_more"`
}
