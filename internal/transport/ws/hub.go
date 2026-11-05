package ws

import (
	"log/slog"
	"sync"
	"uuid"
)

type broadcastMessage struct {
	ListID  uuid.UUID
	Payload []byte
}

type listUserCmd struct {
	userID uuid.UUID
	listID uuid.UUID
}

type Hub struct {
	logger *slog.Logger

	clients       map[*Client]struct{}
	clientsByUser map[uuid.UUID]map[*Client]struct{}
	clientsByList map[uuid.UUID]map[*Client]struct{}

	register   chan *Client
	unregister chan *Client

	addList    chan listUserCmd
	removeList chan listUserCmd
	broadcast  chan broadcastMessage

	done chan struct{}
	once sync.Once
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:        logger,
		clients:       make(map[*Client]struct{}),
		clientsByUser: make(map[uuid.UUID]map[*Client]struct{}),
		clientsByList: make(map[uuid.UUID]map[*Client]struct{}),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		addList:       make(chan listUserCmd),
		removeList:    make(chan listUserCmd),
		broadcast:     make(chan broadcastMessage, 256),
		done:          make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.registerClient(c)

		case c := <-h.unregister:
			h.unregisterClient(c)

		case cmd := <-h.addList:
			h.addListForUser(cmd.userID, cmd.listID)

		case cmd := <-h.removeList:
			h.removeListForUser(cmd.userID, cmd.listID)

		case msg := <-h.broadcast:
			h.dispatch(msg)

		case <-h.done:
			h.shutdown()
			return
		}
	}
}

func (h *Hub) registerClient(c *Client) {
	h.clients[c] = struct{}{}

	clients, ok := h.clientsByUser[c.userID]
	if !ok {
		clients = make(map[*Client]struct{})
		h.clientsByUser[c.userID] = clients
	}
	clients[c] = struct{}{}

	for listID := range c.listIDs {
		h.subscribe(c, listID)
	}
	h.logger.Info("ws client registered",
		"user_id", c.userID, "lists", len(c.listIDs))
}

func (h *Hub) unregisterClient(c *Client) {
	if _, ok := h.clients[c]; !ok {
		return
	}
	delete(h.clients, c)

	if clients, ok := h.clientsByUser[c.userID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.clientsByUser, c.userID)
		}
	}

	for listID := range c.listIDs {
		h.unsubscribe(c, listID)
	}
	close(c.send)
	h.logger.Info("ws client unregistered", "user_id", c.userID)
}

func (h *Hub) addListForUser(userID, listID uuid.UUID) {
	clients, ok := h.clientsByUser[userID]
	if !ok {
		return
	}
	for c := range clients {
		h.subscribe(c, listID)
	}
	h.logger.Debug("ws list added for user", "user_id", userID, "list_id", listID)
}

func (h *Hub) removeListForUser(userID, listID uuid.UUID) {
	clients, ok := h.clientsByUser[userID]
	if !ok {
		return
	}
	for c := range clients {
		h.unsubscribe(c, listID)
	}
	h.logger.Debug("ws list removed for user", "user_id", userID, "list_id", listID)
}

func (h *Hub) subscribe(c *Client, listID uuid.UUID) {
	clients, ok := h.clientsByList[listID]
	if !ok {
		clients = make(map[*Client]struct{})
		h.clientsByList[listID] = clients
	}
	clients[c] = struct{}{}
	c.listIDs[listID] = struct{}{}
}

func (h *Hub) unsubscribe(c *Client, listID uuid.UUID) {
	if clients, ok := h.clientsByList[listID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.clientsByList, listID)
		}
	}
	delete(c.listIDs, listID)
}

func (h *Hub) dispatch(msg broadcastMessage) {
	subs, ok := h.clientsByList[msg.ListID]
	if !ok {
		return
	}

	for c := range subs {
		select {
		case c.send <- msg.Payload:
		default:
			h.logger.Warn("ws client send buffer full, dropping", "user_id", c.userID, "list_id", msg.ListID)
			go func(c *Client) {
				select {
				case h.unregister <- c:
				case <-h.done:
				}
			}(c)
		}
	}
}

func (h *Hub) Register(c *Client) {
	select {
	case h.register <- c:
	case <-h.done:
		c.conn.Close()
		return
	}
}

func (h *Hub) Unregister(c *Client) {
	select {
	case h.unregister <- c:
	case <-h.done:
	}
}

func (h *Hub) RegisterListForUser(userID, listID uuid.UUID) {
	select {
	case h.addList <- listUserCmd{userID: userID, listID: listID}:
	case <-h.done:
	}
}

func (h *Hub) UnregisterListForUser(userID, listID uuid.UUID) {
	select {
	case h.removeList <- listUserCmd{userID: userID, listID: listID}:
	case <-h.done:
	}
}

func (h *Hub) Broadcast(listID uuid.UUID, payload []byte) {
	select {
	case h.broadcast <- broadcastMessage{ListID: listID, Payload: payload}:
	case <-h.done:
	}
}

func (h *Hub) Stop() {
	h.once.Do(func() {
		close(h.done)
	})
}

func (h *Hub) shutdown() {
	for c := range h.clients {
		close(c.send)
	}

	h.clients = make(map[*Client]struct{})
	h.clientsByUser = make(map[uuid.UUID]map[*Client]struct{})
	h.clientsByList = make(map[uuid.UUID]map[*Client]struct{})
}
