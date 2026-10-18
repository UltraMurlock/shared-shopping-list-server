package repository

import (
	"context"
	"errors"
	"fmt"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/shared/database"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type ListRepo struct{}

func NewListRepo() *ListRepo {
	return &ListRepo{}
}

func (r *ListRepo) ListActive(
	ctx context.Context,
	q database.Querier,
) ([]domain.List, error) {
	const query = `
		SELECT id, owner_id, name, description, version, created_at, updated_at, deleted_at
		FROM lists
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list repo: list active: %w", err)
	}
	defer rows.Close()

	lists := make([]domain.List, 0)
	for rows.Next() {
		var l domain.List
		err := rows.Scan(
			&l.ID,
			&l.OwnerID,
			&l.Name,
			&l.Description,
			&l.Version,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("list repo: list active: scan: %w", err)
		}

		lists = append(lists, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list repo: list active: rows: %w", err)
	}

	return lists, nil
}

func (r *ListRepo) GetByID(
	ctx context.Context,
	q database.Querier,
	id uuid.UUID,
) (*domain.List, error) {
	const query = `
		SELECT id, owner_id, name, description, version, created_at, updated_at, deleted_at
		FROM lists
		WHERE id = $1 AND deleted_at IS NULL
	`

	var l domain.List
	err := q.QueryRow(ctx, query, id).Scan(
		&l.ID,
		&l.OwnerID,
		&l.Name,
		&l.Description,
		&l.Version,
		&l.CreatedAt,
		&l.UpdatedAt,
		&l.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("list repo: get by id: %w", domain.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("list repo: get by id: %w", err)
	}

	return &l, nil
}

func (r *ListRepo) Create(
	ctx context.Context,
	q database.Querier,
	ownerId uuid.UUID,
	name, description string,
) (*domain.List, error) {
	const query = `
		INSERT INTO lists (owner_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, owner_id, name, description, version, created_at, updated_at, deleted_at
	`

	var l domain.List
	err := q.QueryRow(ctx, query, ownerId, name, description).Scan(
		&l.ID,
		&l.OwnerID,
		&l.Name,
		&l.Description,
		&l.Version,
		&l.CreatedAt,
		&l.UpdatedAt,
		&l.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("list repo: create: %w", err)
	}

	return &l, nil
}

func (r *ListRepo) Update(
	ctx context.Context,
	q database.Querier,
	id uuid.UUID,
	name, description string,
	expectedVersion int64,
) (*domain.List, error) {
	const query = `
		UPDATE lists
		SET name = $1,
			description = $2,
			version = version + 1,
			updated_at = now()
		WHERE id = $3 AND version = $4 AND deleted_at IS NULL
		RETURNING id, owner_id, name, description, version, created_at, updated_at, deleted_at
	`

	var l domain.List
	err := q.QueryRow(ctx, query, name, description, id, expectedVersion).Scan(
		&l.ID,
		&l.OwnerID,
		&l.Name,
		&l.Description,
		&l.Version,
		&l.CreatedAt,
		&l.UpdatedAt,
		&l.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("list repo: update: %w", domain.ErrVersionConflict)
	}

	if err != nil {
		return nil, fmt.Errorf("list repo: update: %w", err)
	}

	return &l, nil
}

func (r *ListRepo) SoftDelete(
	ctx context.Context,
	q database.Querier,
	id uuid.UUID,
	expectedVersion int64,
) (int64, error) {
	const query = `
		UPDATE lists
		SET deleted_at = now(),
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1 AND version = $2 AND deleted_at IS NULL
		RETURNING version
	`

	var newVersion int64
	err := q.QueryRow(ctx, query, id, expectedVersion).Scan(&newVersion)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("list repo: soft delete: %w", domain.ErrVersionConflict)
	}

	if err != nil {
		return 0, fmt.Errorf("list repo: soft delete: %w", err)
	}

	return newVersion, nil
}
