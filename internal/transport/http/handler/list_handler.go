package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/service"
	"shared-shopping-list-server/internal/transport/http/httputils"
	"shared-shopping-list-server/internal/transport/http/middleware"
	"uuid"
)

type ListHandler struct {
	log         *slog.Logger
	listService *service.ListService
}

func NewListHandler(log *slog.Logger, listService *service.ListService) *ListHandler {
	return &ListHandler{log: log, listService: listService}
}

// List returns all non-deleted shopping lists.
// @Tags        Lists
// @Produce     json
// @Success  	200 {object} listCollectionResponse "Lists collection"
// @Security    BearerAuth
// @Router      /lists [get]
func (h *ListHandler) List(w http.ResponseWriter, r *http.Request) {
	lists, err := h.listService.List(r.Context())
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	resp := make([]listResponse, len(lists))
	for i := range lists {
		resp[i] = toListResponse(&lists[i])
	}

	httputils.WriteJSON(w, http.StatusOK, listCollectionResponse{Lists: resp})
}

// GetByID returns non-deleted shopping list by id.
// @Tags        Lists
// @Produce     json
// @Param      	list_id path string true "List ID (UUID)"
// @Success  	200 {object} listResponse "List"
// @Security    BearerAuth
// @Router      /lists/{list_id} [get]
func (h *ListHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("list_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid list_id")
		return
	}

	list, err := h.listService.GetByID(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	httputils.WriteJSON(w, http.StatusOK, toListResponse(list))
}

// Create creates new shopping list.
// @Tags        Lists
// @Accept		json
// @Produce     json
// @Param     	request body createListRequest true "New list data"
// @Success  	200 {object} listResponse "Created list"
// @Security    BearerAuth
// @Router      /lists [post]
func (h *ListHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputils.WriteError(w, http.StatusInternalServerError, "internal_error", "")
		return
	}

	var req createListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json")
		return
	}

	list, err := h.listService.Create(r.Context(), userID, service.CreateListInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	httputils.WriteJSON(w, http.StatusCreated, toListResponse(list))
}

// Update updates a shopping list by id.
// @Tags        Lists
// @Accept		json
// @Produce     json
// @Param      	list_id path string true "List ID (UUID)"
// @Param     	request body updateListRequest true "Patch"
// @Success  	200 {object} listResponse "Updated list"
// @Security    BearerAuth
// @Router      /lists/{list_id} [patch]
func (h *ListHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	id, err := uuid.Parse(r.PathValue("list_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid list_id")
		return
	}

	var req updateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json")
		return
	}

	list, err := h.listService.Update(r.Context(), userID, id, service.UpdateListInput{
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	httputils.WriteJSON(w, http.StatusOK, toListResponse(list))
}

// Delete soft-deletes a shopping list by id.
// @Tags        Lists
// @Produce     json
// @Param      	list_id path string true "List ID (UUID)"
// @Param     	request body deleteListRequest true "List version"
// @Success  	204 "List deleted"
// @Security    BearerAuth
// @Router      /lists/{list_id} [delete]
func (h *ListHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	id, err := uuid.Parse(r.PathValue("list_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid list_id")
		return
	}

	var req deleteListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json")
		return
	}

	if err := h.listService.SoftDelete(r.Context(), userID, id, req.Version); err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) writeServiceError(w http.ResponseWriter, err error) {
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
