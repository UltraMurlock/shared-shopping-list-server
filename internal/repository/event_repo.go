package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"shared-shopping-list-server/internal/database"
	"shared-shopping-list-server/internal/domain"
	"uuid"
)

type EventRepo struct{}

func NewEventRepo() *EventRepo {
	return &EventRepo{}
}

type NewEvent struct {
	ListID        uuid.UUID
	ActorID       uuid.UUID
	EntityType    domain.EntityType
	EntityID      uuid.UUID
	Operation     domain.Operation
	EntityVersion int64
	Payload       json.RawMessage
}

func (r *EventRepo) Insert(
	ctx context.Context,
	q database.Querier,
	e NewEvent,
) (int64, error) {
	const query = `
		INSERT INTO list_events
			(list_id, actor_id, entity_type, entity_id, operation, entity_version, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int64
	err := q.QueryRow(ctx, query,
		e.ListID,
		e.ActorID,
		e.EntityType,
		e.EntityID,
		e.Operation,
		e.EntityVersion,
		e.Payload,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("event repo: insert: %w", err)
	}

	return id, nil
}

func (r *EventRepo) ListAfter(
	ctx context.Context,
	q database.Querier,
	listID uuid.UUID,
	afterID int64,
	limit int,
) ([]domain.ListEvent, bool, error) {
	const query = `
		SELECT id, list_id, actor_id, entity_type, entity_id, operation, entity_version, payload, created_at
		FROM list_events
		WHERE list_id = $1 AND id > $2
		ORDER BY id
		LIMIT $3
	`

	rows, err := q.Query(ctx, query, listID, afterID, limit+1)
	if err != nil {
		return nil, false, fmt.Errorf("event repo: list after: %w", err)
	}
	defer rows.Close()

	events := make([]domain.ListEvent, 0, limit)
	for rows.Next() {
		var e domain.ListEvent
		var payload []byte
		err := rows.Scan(
			&e.ID,
			&e.ListID,
			&e.ActorID,
			&e.EntityType,
			&e.EntityID,
			&e.Operation,
			&e.EntityVersion,
			&payload,
			&e.CreatedAt,
		)

		if err != nil {
			return nil, false, fmt.Errorf("event repo: list after: scan: %w", err)
		}

		e.Payload = json.RawMessage(payload)
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("event repo: list after: rows: %w", err)
	}

	hasMore := len(events) > limit
	if hasMore {
		events = events[:limit]
	}

	return events, hasMore, nil
}
