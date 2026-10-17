package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"shared-shopping-list-server/internal/database"
	"shared-shopping-list-server/internal/domain"
	"uuid"

	"github.com/jackc/pgx/v5"
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

func (r *EventRepo) ListAfter(
	ctx context.Context,
	q database.Querier,
	afterID int64,
	limit int,
) ([]domain.ListEvent, bool, error) {
	const query = `
		SELECT id, list_id, actor_id, entity_type, entity_id, operation, entity_version, payload, created_at
		FROM list_events
		WHERE id > $1
		ORDER BY id
		LIMIT $2
	`

	rows, err := q.Query(ctx, query, afterID, limit+1)
	if err != nil {
		return nil, false, fmt.Errorf("event repo: list after: %w", err)
	}
	defer rows.Close()

	events := make([]domain.ListEvent, 0, limit)
	for rows.Next() {
		it, err := scanListEvent(rows)
		if err != nil {
			return nil, false, fmt.Errorf("event repo: list after: scan: %w", err)
		}

		events = append(events, *it)
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

func (r *EventRepo) ListByListAfter(
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
		return nil, false, fmt.Errorf("event repo: list by list after: %w", err)
	}
	defer rows.Close()

	events := make([]domain.ListEvent, 0, limit)
	for rows.Next() {
		it, err := scanListEvent(rows)
		if err != nil {
			return nil, false, fmt.Errorf("event repo: list by list after: scan: %w", err)
		}

		events = append(events, *it)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("event repo: list by list after: rows: %w", err)
	}

	hasMore := len(events) > limit
	if hasMore {
		events = events[:limit]
	}

	return events, hasMore, nil
}

func (r *EventRepo) MaxID(ctx context.Context, q database.Querier) (int64, error) {
	const query = `SELECT COALESCE(MAX(id), 0) FROM list_events`

	var id int64
	if err := q.QueryRow(ctx, query).Scan(&id); err != nil {
		return 0, fmt.Errorf("event repo: max id: %w", err)
	}

	return id, nil
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

func scanListEvent(row pgx.Row) (*domain.ListEvent, error) {
	var it domain.ListEvent
	var payload []byte
	err := row.Scan(
		&it.ID,
		&it.ListID,
		&it.ActorID,
		&it.EntityType,
		&it.EntityID,
		&it.Operation,
		&it.EntityVersion,
		&payload,
		&it.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	it.Payload = json.RawMessage(payload)

	return &it, nil
}
