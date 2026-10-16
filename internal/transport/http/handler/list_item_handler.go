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

type ListItemHandler struct {
	log             *slog.Logger
	listItemService *service.ListItemService
}

func NewListItemHandler(log *slog.Logger, svc *service.ListItemService) *ListItemHandler {
	return &ListItemHandler{log: log, listItemService: svc}
}

// List returns all non-deleted list items.
// @Tags        ListItems
// @Produce     json
// @Param      	list_id path string true "List ID (UUID)"
// @Success  	200 {object} listItemCollectionResponse "List items collection"
// @Security    BearerAuth
// @Router      /lists/{list_id}/items [get]
func (h *ListItemHandler) List(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(r.PathValue("list_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid list_id")
		return
	}

	items, err := h.listItemService.ListByList(r.Context(), listID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	resp := make([]listItemResponse, len(items))
	for i := range items {
		resp[i] = toListItemResponse(&items[i])
	}

	httputils.WriteJSON(w, http.StatusOK, listItemCollectionResponse{Items: resp})
}

// Create creates new list item.
// @Tags        ListItems
// @Accept		json
// @Produce     json
// @Param      	list_id path string true "List ID (UUID)"
// @Param     	request body createListItemRequest true "New list item data"
// @Success  	200 {object} listItemResponse "Created list item"
// @Security    BearerAuth
// @Router      /lists/{list_id}/items [post]
func (h *ListItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	listID, err := uuid.Parse(r.PathValue("list_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid list_id")
		return
	}

	var req createListItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json")
		return
	}

	it, err := h.listItemService.Create(r.Context(), userID, listID, service.CreateListItemInput{
		Name:     req.Name,
		Comment:  req.Comment,
		UnitID:   req.UnitID,
		Quantity: req.Quantity,
		IsDone:   req.IsDone,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	httputils.WriteJSON(w, http.StatusCreated, toListItemResponse(it))
}

// Update updates a list item by id.
// @Tags        ListItems
// @Accept		json
// @Produce     json
// @Param      	list_id path string true "List ID (UUID)"
// @Param      	item_id path string true "Item ID (UUID)"
// @Param     	request body updateListItemRequest true "Patch"
// @Success  	200 {object} listItemResponse "Updated list item"
// @Security    BearerAuth
// @Router      /lists/{list_id}/items/{item_id} [patch]
func (h *ListItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	listID, err := uuid.Parse(r.PathValue("list_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid list_id")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("item_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid item_id")
		return
	}

	var req updateListItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json")
		return
	}

	it, err := h.listItemService.Update(r.Context(), userID, listID, itemID, service.UpdateListItemInput{
		Name:     req.Name,
		Comment:  req.Comment,
		UnitID:   req.UnitID,
		Quantity: req.Quantity,
		IsDone:   req.IsDone,
		Version:  req.Version,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	httputils.WriteJSON(w, http.StatusOK, toListItemResponse(it))
}

// Delete soft-deletes a list item by id.
// @Tags        ListItems
// @Produce     json
// @Param      	list_id path string true "List ID (UUID)"
// @Param      	item_id path string true "Item ID (UUID)"
// @Param     	request body deleteListItemRequest true "List item version"
// @Success  	204 "List item deleted"
// @Security    BearerAuth
// @Router      /lists/{list_id}/items/{item_id} [delete]
func (h *ListItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	listID, err := uuid.Parse(r.PathValue("list_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid list_id")
		return
	}
	itemID, err := uuid.Parse(r.PathValue("item_id"))
	if err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid item_id")
		return
	}

	var req deleteListItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json")
		return
	}

	if err := h.listItemService.SoftDelete(r.Context(), userID, listID, itemID, req.Version); err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListItemHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		httputils.WriteError(w, http.StatusNotFound, "not_found", "")
	case errors.Is(err, domain.ErrVersionConflict):
		httputils.WriteError(w, http.StatusConflict, "version_conflict", "")
	case errors.Is(err, domain.ErrValidationError):
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "")
	default:
		h.log.Error("list item handler", "error", err)
		httputils.WriteError(w, http.StatusInternalServerError, "internal_error", "")
	}
}
