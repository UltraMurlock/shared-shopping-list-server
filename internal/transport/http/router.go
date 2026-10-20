package http

import (
	"log/slog"
	"net/http"
	"shared-shopping-list-server/internal/shared/auth"
	"shared-shopping-list-server/internal/transport/http/handler"
	"shared-shopping-list-server/internal/transport/http/middleware"
)

func NewRouter(
	log *slog.Logger,
	jwt *auth.JWTManager,
	corsCfg middleware.CORSConfig,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	unitHandler *handler.UnitHandler,
	listHandler *handler.ListHandler,
	listItemHandler *handler.ListItemHandler,
	eventHandler *handler.EventHandler,
	wsHandler *handler.WSHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.HandleFunc("GET /ws", wsHandler.Connect)

	authMW := middleware.Auth(jwt)
	mux.Handle("GET /users/me", authMW(http.HandlerFunc(userHandler.GetMe)))
	mux.Handle("PATCH /users/me", authMW(http.HandlerFunc(userHandler.UpdateMe)))

	mux.Handle("GET /units", authMW(http.HandlerFunc(unitHandler.List)))

	mux.Handle("GET /lists", authMW(http.HandlerFunc(listHandler.List)))
	mux.Handle("POST /lists", authMW(http.HandlerFunc(listHandler.Create)))
	mux.Handle("GET /lists/{list_id}", authMW(http.HandlerFunc(listHandler.GetByID)))
	mux.Handle("PATCH /lists/{list_id}", authMW(http.HandlerFunc(listHandler.Update)))
	mux.Handle("DELETE /lists/{list_id}", authMW(http.HandlerFunc(listHandler.Delete)))
	mux.Handle("GET /lists/{list_id}/events", authMW(http.HandlerFunc(eventHandler.ListForList)))

	mux.Handle("GET /lists/{list_id}/items", authMW(http.HandlerFunc(listItemHandler.List)))
	mux.Handle("POST /lists/{list_id}/items", authMW(http.HandlerFunc(listItemHandler.Create)))
	mux.Handle("PATCH /lists/{list_id}/items/{item_id}", authMW(http.HandlerFunc(listItemHandler.Update)))
	mux.Handle("DELETE /lists/{list_id}/items/{item_id}", authMW(http.HandlerFunc(listItemHandler.Delete)))

	var h http.Handler = mux
	h = middleware.Recovery(log)(h)
	h = middleware.CORS(corsCfg)(h)
	return h
}
