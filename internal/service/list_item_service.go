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
	"github.com/shopspring/decimal"
)

type ListItemService struct {
	pool         *pgxpool.Pool
	listRepo     *repository.ListRepo
	listItemRepo *repository.ListItemRepo
	unitRepo     *repository.UnitRepo
	eventRepo    *repository.EventRepo
}

func NewListItemService(
	pool *pgxpool.Pool,
	listRepo *repository.ListRepo,
	listItemRepo *repository.ListItemRepo,
	unitRepo *repository.UnitRepo,
	eventRepo *repository.EventRepo,
) *ListItemService {
	return &ListItemService{
		pool:         pool,
		listRepo:     listRepo,
		listItemRepo: listItemRepo,
		unitRepo:     unitRepo,
		eventRepo:    eventRepo,
	}
}

type CreateListItemInput struct {
	Name     string
	Comment  string
	UnitID   int16
	Quantity decimal.Decimal
	IsDone   bool
}

type UpdateListItemInput struct {
	Name     *string
	Comment  *string
	UnitID   *int16
	Quantity *decimal.Decimal
	IsDone   *bool
	Version  int64
}

const (
	maxItemNameLen    = 200
	maxItemCommentLen = 1000
)

func (s *ListItemService) ListByList(
	ctx context.Context,
	listID uuid.UUID,
) ([]domain.ListItem, error) {
	if _, err := s.listRepo.GetByID(ctx, s.pool, listID); err != nil {
		return nil, fmt.Errorf("list item service: list: %w", err)
	}

	items, err := s.listItemRepo.ListByList(ctx, s.pool, listID)
	if err != nil {
		return nil, fmt.Errorf("list item service: list: %w", err)
	}

	return items, nil
}

func (s *ListItemService) GetByID(
	ctx context.Context,
	listID, itemID uuid.UUID,
) (*domain.ListItem, error) {
	it, err := s.listItemRepo.GetByID(ctx, s.pool, listID, itemID)
	if err != nil {
		return nil, fmt.Errorf("list item service: get by id: %w", err)
	}

	return it, nil
}

func (s *ListItemService) Create(
	ctx context.Context,
	actorID uuid.UUID,
	listID uuid.UUID,
	input CreateListItemInput,
) (*domain.ListItem, error) {
	name := strings.TrimSpace(input.Name)
	comment := strings.TrimSpace(input.Comment)

	if err := validateItemName(name); err != nil {
		return nil, fmt.Errorf("list item service: create: %w", err)
	}
	if err := validateItemComment(comment); err != nil {
		return nil, fmt.Errorf("list item service: create: %w", err)
	}
	if err := validateQuantity(input.Quantity); err != nil {
		return nil, fmt.Errorf("list item service: create: %w", err)
	}

	var result *domain.ListItem

	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := s.listRepo.GetByID(ctx, tx, listID); err != nil {
			return err
		}

		if err := s.ensureUnitExists(ctx, tx, input.UnitID); err != nil {
			return err
		}

		it, err := s.listItemRepo.Create(ctx, tx, repository.CreateListItemParams{
			ListID:    listID,
			Name:      name,
			Comment:   comment,
			UnitID:    input.UnitID,
			Quantity:  input.Quantity,
			IsDone:    input.IsDone,
			CreatedBy: actorID,
		})
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"after": map[string]any{
				"name":     it.Name,
				"comment":  it.Comment,
				"unit_id":  it.UnitID,
				"quantity": it.Quantity,
				"is_done":  it.IsDone,
			},
		})
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}

		_, err = s.eventRepo.Insert(ctx, tx, repository.NewEvent{
			ListID:        listID,
			ActorID:       actorID,
			EntityType:    domain.EntityTypeItem,
			EntityID:      it.ID,
			Operation:     domain.OperationCreated,
			EntityVersion: it.Version,
			Payload:       payload,
		})
		if err != nil {
			return err
		}

		result = it
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list item service: create: %w", err)
	}

	return result, nil
}

func (s *ListItemService) Update(
	ctx context.Context,
	actorID uuid.UUID,
	listID, itemID uuid.UUID,
	input UpdateListItemInput,
) (*domain.ListItem, error) {
	if input.Version <= 0 {
		return nil, fmt.Errorf("list item service: update: version: %w", domain.ErrValidationError)
	}
	if input.Name == nil && input.Comment == nil && input.UnitID == nil &&
		input.Quantity == nil && input.IsDone == nil {
		return nil, fmt.Errorf("list item service: update: no fields: %w", domain.ErrValidationError)
	}

	var result *domain.ListItem

	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := s.listRepo.GetByID(ctx, tx, listID); err != nil {
			return err
		}

		current, err := s.listItemRepo.GetByID(ctx, tx, listID, itemID)
		if err != nil {
			return err
		}

		next := *current

		if input.Name != nil {
			name := strings.TrimSpace(*input.Name)
			if err := validateItemName(name); err != nil {
				return err
			}
			next.Name = name
		}
		if input.Comment != nil {
			comment := strings.TrimSpace(*input.Comment)
			if err := validateItemComment(comment); err != nil {
				return err
			}
			next.Comment = comment
		}
		if input.UnitID != nil {
			if err := s.ensureUnitExists(ctx, tx, *input.UnitID); err != nil {
				return err
			}
			next.UnitID = *input.UnitID
		}
		if input.Quantity != nil {
			if err := validateQuantity(*input.Quantity); err != nil {
				return err
			}
			next.Quantity = *input.Quantity
		}
		if input.IsDone != nil {
			next.IsDone = *input.IsDone
		}

		updated, err := s.listItemRepo.Update(ctx, tx, repository.UpdateListItemParams{
			ListID:          listID,
			ItemID:          itemID,
			Name:            next.Name,
			Comment:         next.Comment,
			UnitID:          next.UnitID,
			Quantity:        next.Quantity,
			IsDone:          next.IsDone,
			UpdatedBy:       actorID,
			ExpectedVersion: input.Version,
		})
		if err != nil {
			return err
		}

		payload, err := buildItemUpdatePayload(current, &next)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}

		_, err = s.eventRepo.Insert(ctx, tx, repository.NewEvent{
			ListID:        listID,
			ActorID:       actorID,
			EntityType:    domain.EntityTypeItem,
			EntityID:      itemID,
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
		return nil, fmt.Errorf("list item service: update: %w", err)
	}
	return result, nil
}

func (s *ListItemService) SoftDelete(
	ctx context.Context,
	actorID uuid.UUID,
	listID, itemID uuid.UUID,
	expectedVersion int64,
) error {
	if expectedVersion <= 0 {
		return fmt.Errorf("list item service: soft delete: version: %w", domain.ErrValidationError)
	}

	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := s.listRepo.GetByID(ctx, tx, listID); err != nil {
			return err
		}

		current, err := s.listItemRepo.GetByID(ctx, tx, listID, itemID)
		if err != nil {
			return err
		}

		newVersion, err := s.listItemRepo.SoftDelete(ctx, tx, listID, itemID, actorID, expectedVersion)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"before": map[string]any{
				"name":     current.Name,
				"quantity": current.Quantity,
			},
		})
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}

		_, err = s.eventRepo.Insert(ctx, tx, repository.NewEvent{
			ListID:        listID,
			ActorID:       actorID,
			EntityType:    domain.EntityTypeItem,
			EntityID:      itemID,
			Operation:     domain.OperationDeleted,
			EntityVersion: newVersion,
			Payload:       payload,
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("list item service: soft delete: %w", err)
	}
	return nil
}

func (s *ListItemService) ensureUnitExists(
	ctx context.Context,
	q database.Querier,
	unitID int16,
) error {
	_, err := s.unitRepo.GetByID(ctx, q, unitID)
	if err != nil {
		return fmt.Errorf("check unit: %w", err)
	}

	return nil
}

func buildItemUpdatePayload(before, after *domain.ListItem) ([]byte, error) {
	changed := make([]string, 0)
	b := make(map[string]any)
	a := make(map[string]any)

	if before.Name != after.Name {
		changed = append(changed, "name")
		b["name"] = before.Name
		a["name"] = after.Name
	}
	if before.Comment != after.Comment {
		changed = append(changed, "comment")
		b["comment"] = before.Comment
		a["comment"] = after.Comment
	}
	if before.UnitID != after.UnitID {
		changed = append(changed, "unit_id")
		b["unit_id"] = before.UnitID
		a["unit_id"] = after.UnitID
	}
	if !before.Quantity.Equal(after.Quantity) {
		changed = append(changed, "quantity")
		b["quantity"] = before.Quantity
		a["quantity"] = after.Quantity
	}
	if before.IsDone != after.IsDone {
		changed = append(changed, "is_done")
		b["is_done"] = before.IsDone
		a["is_done"] = after.IsDone
	}

	return json.Marshal(map[string]any{
		"changed": changed,
		"before":  b,
		"after":   a,
	})
}

func validateItemName(name string) error {
	if name == "" {
		return fmt.Errorf("name is empty: %w", domain.ErrValidationError)
	}
	if len([]rune(name)) > maxItemNameLen {
		return fmt.Errorf("name too long: %w", domain.ErrValidationError)
	}

	return nil
}

func validateItemComment(comment string) error {
	if len([]rune(comment)) > maxItemCommentLen {
		return fmt.Errorf("comment too long: %w", domain.ErrValidationError)
	}

	return nil
}

func validateQuantity(q decimal.Decimal) error {
	if !q.IsPositive() {
		return fmt.Errorf("quantity must be > 0: %w", domain.ErrValidationError)
	}

	return nil
}
