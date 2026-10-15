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
	listHandler *handler.ListHandler,
	eventHandler *handler.EventHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", authHandler.Login)

	authMW := middleware.Auth(jwt)
	mux.Handle("GET /users/me", authMW(http.HandlerFunc(userHandler.GetMe)))
	mux.Handle("PATCH /users/me", authMW(http.HandlerFunc(userHandler.UpdateMe)))

	mux.Handle("GET /lists", authMW(http.HandlerFunc(listHandler.List)))
	mux.Handle("POST /lists", authMW(http.HandlerFunc(listHandler.Create)))
	mux.Handle("GET /lists/{list_id}", authMW(http.HandlerFunc(listHandler.GetByID)))
	mux.Handle("PATCH /lists/{list_id}", authMW(http.HandlerFunc(listHandler.Update)))
	mux.Handle("DELETE /lists/{list_id}", authMW(http.HandlerFunc(listHandler.Delete)))
	mux.Handle("GET /lists/{list_id}/events", authMW(http.HandlerFunc(eventHandler.ListForList)))

	var h http.Handler = mux
	recoveryMW := middleware.Recovery(log)
	return recoveryMW(h)
}
