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

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) GetByID(
	ctx context.Context,
	q database.Querier,
	id uuid.UUID,
) (*domain.User, error) {
	const query = `
		SELECT id, login, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var u domain.User
	err := q.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Login,
		&u.DisplayName,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user repo: get by id: %w", domain.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("user repo: get by id: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) GetByLogin(
	ctx context.Context,
	q database.Querier,
	login string,
) (*domain.User, error) {
	const query = `
		SELECT id, login, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE login = $1
	`

	var u domain.User
	err := q.QueryRow(ctx, query, login).Scan(
		&u.ID,
		&u.Login,
		&u.DisplayName,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user repo: get by login: %w", domain.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("user repo: get by login: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) UpdateDisplayName(
	ctx context.Context,
	q database.Querier,
	id uuid.UUID,
	displayName string,
) (*domain.User, error) {
	const query = `
		UPDATE users
		SET display_name = $1, updated_at = now()
		WHERE id = $2
		RETURNING id, login, display_name, password_hash, created_at, updated_at
	`

	var u domain.User
	err := q.QueryRow(ctx, query, displayName, id).Scan(
		&u.ID,
		&u.Login,
		&u.DisplayName,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user repo: update display name: %w", domain.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("user repo: update display name: %w", err)
	}

	return &u, nil
}
