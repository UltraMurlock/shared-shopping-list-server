package ws

import (
	"encoding/json"
	"time"
	"uuid"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID uuid.UUID
	send   chan []byte

	subscriptions map[uuid.UUID]struct{}
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		userID:        userID,
		send:          make(chan []byte, 256),
		subscriptions: make(map[uuid.UUID]struct{}),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error("ws read error", "error", err, "user_id", c.userID)
			}
			break
		}

		var cmd command
		if err := json.Unmarshal(message, &cmd); err != nil {
			c.hub.logger.Warn("ws invalid command", "error", err, "user_id", c.userID)
			continue
		}

		switch cmd.Type {
		case "subscribe":
			listID, err := uuid.Parse(cmd.ListID)
			if err != nil {
				c.hub.logger.Warn("ws invalid list_id", "error", err, "user_id", c.userID)
				continue
			}
			c.subscriptions[listID] = struct{}{}
			c.hub.Subscribe(c, listID)

		case "unsubscribe":
			listID, err := uuid.Parse(cmd.ListID)
			if err != nil {
				continue
			}
			delete(c.subscriptions, listID)
			c.hub.Unsubscribe(c, listID)

		default:
			c.hub.logger.Warn("ws unknown command type", "type", cmd.Type, "user_id", c.userID)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type command struct {
	Type   string `json:"type"`
	ListID string `json:"list_id"`
}
