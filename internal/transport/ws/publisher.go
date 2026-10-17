package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/repository"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	publisherBatchSize = 100
	publisherInterval  = time.Second
)

type Publisher struct {
	logger    *slog.Logger
	pool      *pgxpool.Pool
	eventRepo *repository.EventRepo
	hub       *Hub

	lastEventID int64
}

func NewPublisher(
	logger *slog.Logger,
	pool *pgxpool.Pool,
	eventRepo *repository.EventRepo,
	hub *Hub,
) *Publisher {
	return &Publisher{
		logger:    logger,
		pool:      pool,
		eventRepo: eventRepo,
		hub:       hub,
	}
}

type envelope struct {
	Type   string           `json:"type"`
	ListID uuid.UUID        `json:"list_id"`
	Event  domain.ListEvent `json:"event"`
}

func (p *Publisher) Run(ctx context.Context) error {
	maxID, err := p.eventRepo.MaxID(ctx, p.pool)
	if err != nil {
		return fmt.Errorf("publisher: init cursor: %w", err)
	}

	p.lastEventID = maxID
	p.logger.Info("publisher started", "last_event_id", p.lastEventID)

	ticker := time.NewTicker(publisherInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("publisher stopped")
			return nil

		case <-ticker.C:
			if err := p.poll(ctx); err != nil {
				p.logger.Error("publisher poll", "error", err)
			}
		}
	}
}

func (p *Publisher) poll(ctx context.Context) error {
	for {
		events, hasMore, err := p.eventRepo.ListAfter(ctx, p.pool, p.lastEventID, publisherBatchSize)
		if err != nil {
			return fmt.Errorf("poll: %w", err)
		}
		if len(events) == 0 {
			return nil
		}

		for _, e := range events {
			payload, err := json.Marshal(envelope{
				Type:   "event",
				ListID: e.ListID,
				Event:  e,
			})
			if err != nil {
				p.logger.Error("publisher marshal", "error", err, "event_id", e.ID)
				continue
			}

			p.hub.Broadcast(e.ListID, payload)
			p.lastEventID = e.ID
		}

		if !hasMore {
			return nil
		}
	}
}
