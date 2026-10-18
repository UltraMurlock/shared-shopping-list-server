package repository

import (
	"context"
	"errors"
	"fmt"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/shared/database"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type ListItemRepo struct{}

func NewListItemRepo() *ListItemRepo {
	return &ListItemRepo{}
}

const listItemColumns = `
	id, list_id, name, comment, unit_id, quantity, is_done,	version, 
	created_by, created_at, updated_by, updated_at, deleted_by, deleted_at
`

func (r *ListItemRepo) ListByList(
	ctx context.Context,
	q database.Querier,
	listID uuid.UUID,
) ([]domain.ListItem, error) {
	query := `SELECT ` + listItemColumns + `
		FROM list_items
		WHERE list_id = $1 AND deleted_at IS NULL
		ORDER BY created_at
	`

	rows, err := q.Query(ctx, query, listID)
	if err != nil {
		return nil, fmt.Errorf("list item repo: list by list: %w", err)
	}
	defer rows.Close()

	items := make([]domain.ListItem, 0)
	for rows.Next() {
		it, err := scanListItem(rows)
		if err != nil {
			return nil, fmt.Errorf("list item repo: list by list: scan: %w", err)
		}

		items = append(items, *it)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list item repo: list by list: rows: %w", err)
	}

	return items, nil
}

func (r *ListItemRepo) GetByID(
	ctx context.Context,
	q database.Querier,
	listID, itemID uuid.UUID,
) (*domain.ListItem, error) {
	query := `SELECT ` + listItemColumns + `
		FROM list_items
		WHERE id = $1 AND list_id = $2 AND deleted_at IS NULL
	`

	it, err := scanListItem(q.QueryRow(ctx, query, itemID, listID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("list item repo: get by id: %w", domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("list item repo: get by id: %w", err)
	}

	return it, nil
}

type CreateListItemParams struct {
	ListID    uuid.UUID
	Name      string
	Comment   string
	UnitID    int16
	Quantity  decimal.Decimal
	IsDone    bool
	CreatedBy uuid.UUID
}

func (r *ListItemRepo) Create(
	ctx context.Context,
	q database.Querier,
	p CreateListItemParams,
) (*domain.ListItem, error) {
	query := `INSERT INTO list_items
			(list_id, name, comment, unit_id, quantity, is_done, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING` + listItemColumns

	it, err := scanListItem(q.QueryRow(ctx, query,
		p.ListID, p.Name, p.Comment, p.UnitID, p.Quantity, p.IsDone, p.CreatedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("list item repo: create: %w", err)
	}

	return it, nil
}

type UpdateListItemParams struct {
	ListID          uuid.UUID
	ItemID          uuid.UUID
	Name            string
	Comment         string
	UnitID          int16
	Quantity        decimal.Decimal
	IsDone          bool
	UpdatedBy       uuid.UUID
	ExpectedVersion int64
}

func (r *ListItemRepo) Update(
	ctx context.Context,
	q database.Querier,
	p UpdateListItemParams,
) (*domain.ListItem, error) {
	query := `UPDATE list_items
		SET name = $1,
		    comment = $2,
		    unit_id = $3,
		    quantity = $4,
		    is_done = $5,
		    version = version + 1,
		    updated_by = $6,
		    updated_at = now()
		WHERE id = $7 AND list_id = $8 AND version = $9 AND deleted_at IS NULL
		RETURNING ` + listItemColumns

	it, err := scanListItem(q.QueryRow(ctx, query,
		p.Name, p.Comment, p.UnitID, p.Quantity, p.IsDone, p.UpdatedBy, p.ItemID, p.ListID, p.ExpectedVersion,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("list item repo: update: %w", domain.ErrVersionConflict)
	}
	if err != nil {
		return nil, fmt.Errorf("list item repo: update: %w", err)
	}

	return it, nil
}

func (r *ListItemRepo) SoftDelete(
	ctx context.Context,
	q database.Querier,
	listID, itemID, deletedBy uuid.UUID,
	expectedVersion int64,
) (int64, error) {
	const query = `
		UPDATE list_items
		SET deleted_at = now(),
		    deleted_by = $1,
		    version = version + 1,
		    updated_at = now(),
		    updated_by = $1
		WHERE id = $2 AND list_id = $3 AND version = $4 AND deleted_at IS NULL
		RETURNING version
	`

	var newVersion int64
	err := q.QueryRow(ctx, query, deletedBy, itemID, listID, expectedVersion).Scan(&newVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("list item repo: soft delete: %w", domain.ErrVersionConflict)
	}
	if err != nil {
		return 0, fmt.Errorf("list item repo: soft delete: %w", err)
	}

	return newVersion, nil
}

func scanListItem(row pgx.Row) (*domain.ListItem, error) {
	var it domain.ListItem
	err := row.Scan(
		&it.ID,
		&it.ListID,
		&it.Name,
		&it.Comment,
		&it.UnitID,
		&it.Quantity,
		&it.IsDone,
		&it.Version,
		&it.CreatedBy,
		&it.CreatedAt,
		&it.UpdatedBy,
		&it.UpdatedAt,
		&it.DeletedBy,
		&it.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &it, nil
}
