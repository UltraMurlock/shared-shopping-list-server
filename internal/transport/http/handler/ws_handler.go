package handler

import (
	"log/slog"
	"net/http"
	"shared-shopping-list-server/internal/service"
	"shared-shopping-list-server/internal/shared/auth"
	"shared-shopping-list-server/internal/transport/http/httputils"
	"shared-shopping-list-server/internal/transport/ws"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: для MVP; в проде — проверка Origin
	},
}

type WSHandler struct {
	logger      *slog.Logger
	hub         *ws.Hub
	jwt         *auth.JWTManager
	listService *service.ListService
}

func NewWSHandler(
	logger *slog.Logger,
	hub *ws.Hub,
	jwt *auth.JWTManager,
	listService *service.ListService,
) *WSHandler {
	return &WSHandler{logger: logger, hub: hub, jwt: jwt, listService: listService}
}

// Connect upgrades HTTP to WebSocket and authenticates via token query parameter.
// @Summary     WebSocket-connection
// @Tags        WebSocket
// @Param		token query string true "JWT-token"
// @Success     101"Switching Protocols — WebSocket-connection established"
// @Router      /ws [get]
func (h *WSHandler) Connect(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwt.Parse(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	userID, err := claims.UserID()
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	listIDs, err := h.listService.GetUserListIDs(r.Context(), userID)
	if err != nil {
		h.logger.Error("ws load lists", "error", err, "user_id", userID)
		httputils.WriteError(w, http.StatusInternalServerError, "internal_error", "")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade", "error", err)
		return
	}

	client := ws.NewClient(h.hub, conn, userID, listIDs)
	h.hub.Register(client)

	go client.WritePump()
	client.ReadPump()
}
