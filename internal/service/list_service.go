package service

import (
	"context"
	"encoding/json"
	"fmt"
	"shared-shopping-list-server/internal/database"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/repository"
	"strings"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListService struct {
	pool      *pgxpool.Pool
	listRepo  *repository.ListRepo
	eventRepo *repository.EventRepo
}

func NewListService(
	pool *pgxpool.Pool,
	listRepo *repository.ListRepo,
	eventRepo *repository.EventRepo,
) *ListService {
	return &ListService{
		pool:      pool,
		listRepo:  listRepo,
		eventRepo: eventRepo,
	}
}

const (
	maxListNameLen        = 200
	maxListDescriptionLen = 1000
)

type CreateListInput struct {
	Name        string
	Description string
}

type UpdateListInput struct {
	Name        *string
	Description *string
	Version     int64
}

func (s *ListService) List(ctx context.Context) ([]domain.List, error) {
	lists, err := s.listRepo.ListActive(ctx, s.pool)
	if err != nil {
		return nil, fmt.Errorf("list service: list: %w", err)
	}

	return lists, nil
}

func (s *ListService) GetByID(ctx context.Context, id uuid.UUID) (*domain.List, error) {
	l, err := s.listRepo.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, fmt.Errorf("list service: get by id: %w", err)
	}

	return l, nil
}

func (s *ListService) Create(
	ctx context.Context,
	actorID uuid.UUID,
	input CreateListInput,
) (*domain.List, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)

	if err := validateListName(name); err != nil {
		return nil, fmt.Errorf("list service: create: %w", domain.ErrValidationError)
	}
	if err := validateListDescription(description); err != nil {
		return nil, fmt.Errorf("list service: create: %w", domain.ErrValidationError)
	}

	var result *domain.List

	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		l, err := s.listRepo.Create(ctx, tx, actorID, name, description)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"after": map[string]any{
				"name":        l.Name,
				"description": l.Description,
			},
		})
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}

		_, err = s.eventRepo.Insert(ctx, tx, repository.NewEvent{
			ListID:        l.ID,
			ActorID:       actorID,
			EntityType:    domain.EntityTypeList,
			EntityID:      l.ID,
			Operation:     domain.OperationCreated,
			EntityVersion: l.Version,
			Payload:       payload,
		})
		if err != nil {
			return err
		}

		result = l
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list service: create: %w", err)
	}

	return result, nil
}

func (s *ListService) Update(
	ctx context.Context,
	actorID uuid.UUID,
	listID uuid.UUID,
	input UpdateListInput,
) (*domain.List, error) {
	if input.Version <= 0 {
		return nil, fmt.Errorf("list service: update: version: %w", domain.ErrValidationError)
	}

	if input.Name == nil && input.Description == nil {
		return nil, fmt.Errorf("list service: update: no fields: %w", domain.ErrValidationError)
	}

	var result *domain.List

	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		current, err := s.listRepo.GetByID(ctx, tx, listID)
		if err != nil {
			return err
		}

		newName := current.Name
		if input.Name != nil {
			newName = strings.TrimSpace(*input.Name)
			if err := validateListName(newName); err != nil {
				return fmt.Errorf("list service: update: %w", domain.ErrValidationError)
			}
		}

		newDescription := current.Description
		if input.Description != nil {
			newDescription = strings.TrimSpace(*input.Description)
			if err := validateListDescription(newDescription); err != nil {
				return fmt.Errorf("list service: update: %w", domain.ErrValidationError)
			}
		}

		updated, err := s.listRepo.Update(ctx, tx, listID, newName, newDescription, input.Version)
		if err != nil {
			return err
		}

		payload, err := buildUpdatePayload(current, newName, newDescription)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}

		_, err = s.eventRepo.Insert(ctx, tx, repository.NewEvent{
			ListID:        listID,
			ActorID:       actorID,
			EntityType:    domain.EntityTypeList,
			EntityID:      listID,
			Operation:     domain.OperationUpdated,
			EntityVersion: updated.Version,
			Payload:       payload,
		})
		if err != nil {
			return err
		}

		result = updated
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list service: update: %w", err)
	}

	return result, nil
}

func (s *ListService) SoftDelete(
	ctx context.Context,
	actorID uuid.UUID,
	listID uuid.UUID,
	expectedVersion int64,
) error {
	if expectedVersion <= 0 {
		return fmt.Errorf("list service: soft delete: version: %w", domain.ErrValidationError)
	}

	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		current, err := s.listRepo.GetByID(ctx, tx, listID)
		if err != nil {
			return err
		}

		newVersion, err := s.listRepo.SoftDelete(ctx, tx, listID, expectedVersion)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"before": map[string]any{
				"name":        current.Name,
				"description": current.Description,
			},
		})
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}

		_, err = s.eventRepo.Insert(ctx, tx, repository.NewEvent{
			ListID:        listID,
			ActorID:       actorID,
			EntityType:    domain.EntityTypeList,
			EntityID:      listID,
			Operation:     domain.OperationDeleted,
			EntityVersion: newVersion,
			Payload:       payload,
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("list service: soft delete: %w", err)
	}

	return nil
}

func buildUpdatePayload(current *domain.List, newName, newDescription string) ([]byte, error) {
	changed := make([]string, 0, 2)
	before := make(map[string]any)
	after := make(map[string]any)

	if newName != current.Name {
		changed = append(changed, "name")
		before["name"] = current.Name
		after["name"] = newName
	}
	if newDescription != current.Description {
		changed = append(changed, "description")
		before["description"] = current.Description
		after["description"] = newDescription
	}

	return json.Marshal(map[string]any{
		"changed": changed,
		"before":  before,
		"after":   after,
	})
}

func validateListName(name string) error {
	if name == "" {
		return fmt.Errorf("name is empty: %w", domain.ErrValidationError)
	}

	if len([]rune(name)) > maxListNameLen {
		return fmt.Errorf("name too long: %w", domain.ErrValidationError)
	}

	return nil
}

func validateListDescription(description string) error {
	if len([]rune(description)) > maxListDescriptionLen {
		return fmt.Errorf("description too long: %w", domain.ErrValidationError)
	}

	return nil
}
