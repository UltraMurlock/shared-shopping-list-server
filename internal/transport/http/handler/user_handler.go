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
)

type UserHandler struct {
	log         *slog.Logger
	userService *service.UserService
}

func NewUserHandler(log *slog.Logger, userService *service.UserService) *UserHandler {
	return &UserHandler{log: log, userService: userService}
}

// GetMe returns the current user's profile.
// @Tags        Users
// @Produce     json
// @Success  	200 {object} userResponse "User"
// @Security    BearerAuth
// @Router      /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	u, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	httputils.WriteJSON(w, http.StatusOK, toUserResponse(u))
}

// UpdateMe updates the current user's profile.
// @Tags        Users
// @Accept		json
// @Produce     json
// @Param     	request body updateMeRequest true "Patch"
// @Success     200 {object} userResponse "User"
// @Security    BearerAuth
// @Router      /users/me [patch]
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json")
		return
	}

	if req.DisplayName == nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "nothing to update")
		return
	}

	u, err := h.userService.UpdateDisplayName(r.Context(), userID, *req.DisplayName)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	httputils.WriteJSON(w, http.StatusOK, toUserResponse(u))
}

func (h *UserHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		httputils.WriteError(w, http.StatusNotFound, "not_found", "")
	case errors.Is(err, domain.ErrValidationError):
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "")
	default:
		h.log.Error("user handler", "error", err)
		httputils.WriteError(w, http.StatusInternalServerError, "internal_error", "")
	}
}
