package handler

import (
	"shared-shopping-list-server/internal/domain"
	"time"
	"uuid"
)

type userResponse struct {
	ID          uuid.UUID `json:"id"`
	Login       string    `json:"login"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:          u.ID,
		Login:       u.Login,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

type updateMeRequest struct {
	DisplayName *string `json:"display_name"`
}
