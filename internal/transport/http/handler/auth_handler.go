package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/service"
	"shared-shopping-list-server/internal/transport/http/httputils"
)

type AuthHandler struct {
	log         *slog.Logger
	authService *service.AuthService
}

func NewAuthHandler(log *slog.Logger, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{log: log, authService: authService}
}

// Login authenticates a user and returns a JWT.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body loginRequest true "Credentials"
// @Success      200 {object} loginResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json")
		return
	}

	result, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		case errors.Is(err, domain.ErrValidationError):
			httputils.WriteError(w, http.StatusBadRequest, "validation_error", "")
		default:
			h.log.Error("login", "error", err)
			httputils.WriteError(w, http.StatusInternalServerError, "internal_error", "")
		}

		return
	}

	httputils.WriteJSON(w, http.StatusOK, loginResponse{
		Token: result.Token,
		User:  toUserResponse(result.User),
	})
}
