package handler

import "shared-shopping-list-server/internal/domain"

type unitResponse struct {
	ID   int16  `json:"id"`
	Name string `json:"name"`
}

func toUnitResponse(u *domain.Unit) unitResponse {
	return unitResponse{
		ID:   u.ID,
		Name: u.Name,
	}
}

type unitCollectionResponse struct {
	Units []unitResponse `json:"units"`
}
