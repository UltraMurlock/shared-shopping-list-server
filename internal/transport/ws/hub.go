package ws

import (
	"log/slog"
	"sync"
	"uuid"
)

type subscribeCmd struct {
	client *Client
	listID uuid.UUID
}

type broadcastMessage struct {
	ListID  uuid.UUID
	Payload []byte
}

type Hub struct {
	logger *slog.Logger

	clients       map[*Client]struct{}
	subscriptions map[uuid.UUID]map[*Client]struct{}

	register   chan *Client
	unregister chan *Client

	subscribe   chan subscribeCmd
	unsubscribe chan subscribeCmd
	broadcast   chan broadcastMessage

	done chan struct{}
	once sync.Once
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:        logger,
		clients:       make(map[*Client]struct{}),
		subscriptions: make(map[uuid.UUID]map[*Client]struct{}),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		subscribe:     make(chan subscribeCmd),
		unsubscribe:   make(chan subscribeCmd),
		broadcast:     make(chan broadcastMessage, 256),
		done:          make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
			h.logger.Info("ws client registered", "user_id", c.userID)

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				h.removeFromAllSubscriptions(c)
				close(c.send)
				h.logger.Info("ws client unregistered", "user_id", c.userID)
			}

		case cmd := <-h.subscribe:
			h.subscriptionsAdd(cmd.client, cmd.listID)

		case cmd := <-h.unsubscribe:
			h.subscriptionsRemove(cmd.client, cmd.listID)

		case msg := <-h.broadcast:
			h.dispatch(msg)

		case <-h.done:
			h.shutdown()
			return
		}
	}
}

func (h *Hub) subscriptionsAdd(c *Client, listID uuid.UUID) {
	subs, ok := h.subscriptions[listID]
	if !ok {
		subs = make(map[*Client]struct{})
		h.subscriptions[listID] = subs
	}

	subs[c] = struct{}{}
	h.logger.Debug("ws subscribed", "user_id", c.userID, "list_id", listID)
}

func (h *Hub) subscriptionsRemove(c *Client, listID uuid.UUID) {
	subs, ok := h.subscriptions[listID]
	if !ok {
		return
	}
	delete(subs, c)
	if len(subs) == 0 {
		delete(h.subscriptions, listID)
	}
	h.logger.Debug("ws unsubscribed", "user_id", c.userID, "list_id", listID)
}

func (h *Hub) removeFromAllSubscriptions(c *Client) {
	for listID := range c.subscriptions {
		h.subscriptionsRemove(c, listID)
	}
}

func (h *Hub) dispatch(msg broadcastMessage) {
	subs, ok := h.subscriptions[msg.ListID]
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

func (h *Hub) Subscribe(c *Client, listID uuid.UUID) {
	select {
	case h.subscribe <- subscribeCmd{client: c, listID: listID}:
	case <-h.done:
	}
}

func (h *Hub) Unsubscribe(c *Client, listID uuid.UUID) {
	select {
	case h.unsubscribe <- subscribeCmd{client: c, listID: listID}:
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
	h.subscriptions = make(map[uuid.UUID]map[*Client]struct{})
}
