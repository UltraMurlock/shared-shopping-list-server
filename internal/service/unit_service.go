package service

import (
	"context"
	"fmt"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitService struct {
	pool     *pgxpool.Pool
	unitRepo *repository.UnitRepo
}

func NewUnitService(
	pool *pgxpool.Pool,
	unitRepo *repository.UnitRepo,
) *UnitService {
	return &UnitService{pool: pool, unitRepo: unitRepo}
}

func (s *UnitService) List(ctx context.Context) ([]domain.Unit, error) {
	units, err := s.unitRepo.List(ctx, s.pool)
	if err != nil {
		return nil, fmt.Errorf("unit service: list: %w", err)
	}

	return units, nil
}
