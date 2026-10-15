package service

import (
	"context"
	"fmt"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/repository"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventService struct {
	pool      *pgxpool.Pool
	listRepo  *repository.ListRepo
	eventRepo *repository.EventRepo
}

func NewEventService(
	pool *pgxpool.Pool,
	listRepo *repository.ListRepo,
	eventRepo *repository.EventRepo,
) *EventService {
	return &EventService{pool: pool, listRepo: listRepo, eventRepo: eventRepo}
}

func (s *EventService) ListForList(
	ctx context.Context,
	listID uuid.UUID,
	afterID int64,
	limit int,
) ([]domain.ListEvent, bool, error) {
	if _, err := s.listRepo.GetByID(ctx, s.pool, listID); err != nil {
		return nil, false, fmt.Errorf("event service: list for list: %w", err)
	}

	events, hasMore, err := s.eventRepo.ListAfter(ctx, s.pool, listID, afterID, limit)
	if err != nil {
		return nil, false, fmt.Errorf("event service: list for list: %w", err)
	}

	return events, hasMore, nil
}
