package handler

import (
	"shared-shopping-list-server/internal/domain"
	"time"
	"uuid"

	"github.com/shopspring/decimal"
)

type listItemResponse struct {
	ID        uuid.UUID       `json:"id"`
	ListID    uuid.UUID       `json:"list_id"`
	Name      string          `json:"name"`
	Comment   string          `json:"comment"`
	UnitID    int16           `json:"unit_id"`
	Quantity  decimal.Decimal `json:"quantity"`
	IsDone    bool            `json:"is_done"`
	Version   int64           `json:"version"`
	CreatedBy uuid.UUID       `json:"created_by"`
	UpdatedBy uuid.UUID       `json:"updated_by"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func toListItemResponse(it *domain.ListItem) listItemResponse {
	return listItemResponse{
		ID:        it.ID,
		ListID:    it.ListID,
		Name:      it.Name,
		Comment:   it.Comment,
		UnitID:    it.UnitID,
		Quantity:  it.Quantity,
		IsDone:    it.IsDone,
		Version:   it.Version,
		CreatedBy: it.CreatedBy,
		UpdatedBy: it.UpdatedBy,
		CreatedAt: it.CreatedAt,
		UpdatedAt: it.UpdatedAt,
	}
}

type listItemCollectionResponse struct {
	Items []listItemResponse `json:"items"`
}

type createListItemRequest struct {
	Name     string          `json:"name"`
	Comment  string          `json:"comment"`
	UnitID   int16           `json:"unit_id"`
	Quantity decimal.Decimal `json:"quantity"`
	IsDone   bool            `json:"is_done"`
}

type updateListItemRequest struct {
	Version  int64            `json:"version"`
	Name     *string          `json:"name"`
	Comment  *string          `json:"comment"`
	UnitID   *int16           `json:"unit_id"`
	Quantity *decimal.Decimal `json:"quantity"`
	IsDone   *bool            `json:"is_done"`
}

type deleteListItemRequest struct {
	Version int64 `json:"version"`
}
