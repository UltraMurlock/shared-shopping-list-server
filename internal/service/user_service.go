package service

import (
	"context"
	"fmt"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/repository"
	"strings"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	pool     *pgxpool.Pool
	userRepo *repository.UserRepo
}

func NewUserService(pool *pgxpool.Pool, userRepo *repository.UserRepo) *UserService {
	return &UserService{pool: pool, userRepo: userRepo}
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.userRepo.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, fmt.Errorf("user service: get by id: %w", err)
	}

	return u, nil
}

func (s *UserService) UpdateDisplayName(
	ctx context.Context,
	id uuid.UUID,
	displayName string,
) (*domain.User, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || len([]rune(displayName)) > 100 {
		return nil, fmt.Errorf("user service: update display name: %w", domain.ErrValidationError)
	}

	u, err := s.userRepo.UpdateDisplayName(ctx, s.pool, id, displayName)
	if err != nil {
		return nil, fmt.Errorf("user service: update display name: %w", err)
	}

	return u, nil
}
