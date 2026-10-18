package repository

import (
	"context"
	"errors"
	"fmt"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/shared/database"

	"github.com/jackc/pgx/v5"
)

type UnitRepo struct{}

func NewUnitRepo() *UnitRepo {
	return &UnitRepo{}
}

func (r *UnitRepo) List(
	ctx context.Context,
	q database.Querier,
) ([]domain.Unit, error) {
	const query = `
		SELECT id, name
		FROM units
		ORDER BY id
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unit repo: list: %w", err)
	}
	defer rows.Close()

	units := make([]domain.Unit, 0)
	for rows.Next() {
		var u domain.Unit
		err := rows.Scan(
			&u.ID,
			&u.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("unit repo: list: scan: %w", err)
		}

		units = append(units, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unit repo: list: rows: %w", err)
	}

	return units, nil
}

func (r *UnitRepo) GetByID(
	ctx context.Context,
	q database.Querier,
	id int16,
) (*domain.Unit, error) {
	const query = `
		SELECT id, name
		FROM units
		WHERE id = $1
	`

	var l domain.Unit
	err := q.QueryRow(ctx, query, id).Scan(&l.ID, &l.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("unit repo: get by id: %w", domain.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("unit repo: get by id: %w", err)
	}

	return &l, nil
}
