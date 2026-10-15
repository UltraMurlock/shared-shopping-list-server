package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/service"
	"shared-shopping-list-server/internal/transport/http/httputils"
	"strconv"
	"uuid"
)

type EventHandler struct {
	log          *slog.Logger
	eventService *service.EventService
}

func NewEventHandler(log *slog.Logger, eventService *service.EventService) *EventHandler {
	return &EventHandler{log: log, eventService: eventService}
}

// List returns all non-deleted shopping lists.
// @Tags        Lists
// @Produce     json
// @Param      	list_id path string true "List ID (UUID)"
// @Param       after query integer true "Cursor: return event with id > after (0 — from start)" minimum(0)
// @Param       limit query integer false "Max batch size" minimum(1) maximum(500)
// @Success  	200 {object} listEventsResponse "Events"
// @Security    BearerAuth
// @Router      /lists/{list_id}/events [get]
func (h *EventHandler) ListForList(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("list_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid list_id")
		return
	}

	afterStr := r.URL.Query().Get("after")
	if afterStr == "" {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "after is required")
		return
	}
	after, err := strconv.ParseInt(afterStr, 10, 64)
	if err != nil || after < 0 {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid after")
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		v, err := strconv.Atoi(limitStr)
		if err != nil || v <= 0 {
			httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid limit")
			return
		}

		limit = v
	}
	if limit > 500 {
		limit = 500
	}

	events, hasMore, err := h.eventService.ListForList(r.Context(), id, after, limit)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	respEvents := make([]eventResponse, len(events))
	for i := range events {
		respEvents[i] = toEventResponse(events[i])
	}

	httputils.WriteJSON(w, http.StatusOK, listEventsResponse{
		Events:  respEvents,
		HasMore: hasMore,
	})
}

func (h *EventHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		httputils.WriteError(w, http.StatusNotFound, "not_found", "")
	case errors.Is(err, domain.ErrVersionConflict):
		httputils.WriteError(w, http.StatusConflict, "version_conflict", "")
	case errors.Is(err, domain.ErrValidationError):
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "")
	default:
		h.log.Error("event handler", "error", err)
		httputils.WriteError(w, http.StatusInternalServerError, "internal_error", "")
	}
}
