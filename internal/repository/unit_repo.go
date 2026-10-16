package repository

import (
	"context"
	"fmt"
	"shared-shopping-list-server/internal/database"
	"shared-shopping-list-server/internal/domain"
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
