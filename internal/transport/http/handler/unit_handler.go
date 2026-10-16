package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/service"
	"shared-shopping-list-server/internal/transport/http/httputils"
)

type UnitHandler struct {
	log         *slog.Logger
	unitService *service.UnitService
}

func NewUnitHandler(log *slog.Logger, unitService *service.UnitService) *UnitHandler {
	return &UnitHandler{log: log, unitService: unitService}
}

// List returns all units.
// @Tags        Units
// @Produce     json
// @Success  	200 {object} unitCollectionResponse "Units"
// @Security    BearerAuth
// @Router      /units [get]
func (h *UnitHandler) List(w http.ResponseWriter, r *http.Request) {
	units, err := h.unitService.List(r.Context())
	if err != nil {
		h.writeServiceError(w, err)
	}

	unitDtos := make([]unitResponse, len(units))
	for i := range units {
		unitDtos[i] = toUnitResponse(&units[i])
	}

	httputils.WriteJSON(w, http.StatusOK, unitCollectionResponse{Units: unitDtos})
}

func (h *UnitHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		httputils.WriteError(w, http.StatusNotFound, "not_found", "")
	case errors.Is(err, domain.ErrVersionConflict):
		httputils.WriteError(w, http.StatusConflict, "version_conflict", "")
	case errors.Is(err, domain.ErrValidationError):
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "")
	default:
		h.log.Error("list handler", "error", err)
		httputils.WriteError(w, http.StatusInternalServerError, "internal_error", "")
	}
}
