package handler

import (
	"shared-shopping-list-server/internal/domain"
	"time"
	"uuid"
)

type listResponse struct {
	ID          uuid.UUID `json:"id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     int64     `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toListResponse(l *domain.List) listResponse {
	return listResponse{
		ID:          l.ID,
		OwnerID:     l.OwnerID,
		Name:        l.Name,
		Description: l.Description,
		Version:     l.Version,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

type listCollectionResponse struct {
	Lists []listResponse `json:"lists"`
}

type createListRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type updateListRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Version     int64   `json:"version" validate:"required"`
}

type deleteListRequest struct {
	Version int64 `json:"version"`
}
