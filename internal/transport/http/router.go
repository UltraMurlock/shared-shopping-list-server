package http

import (
	"log/slog"
	"net/http"
	"shared-shopping-list-server/internal/auth"
	"shared-shopping-list-server/internal/transport/http/handler"
	"shared-shopping-list-server/internal/transport/http/middleware"
)

func NewRouter(
	log *slog.Logger,
	jwt *auth.JWTManager,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", authHandler.Login)

	authMW := middleware.Auth(jwt)
	mux.Handle("GET /users/me", authMW(http.HandlerFunc(userHandler.GetMe)))
	mux.Handle("PATCH /users/me", authMW(http.HandlerFunc(userHandler.UpdateMe)))

	var h http.Handler = mux
	recoveryMW := middleware.Recovery(log)
	return recoveryMW(h)
}
