package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu       sync.Mutex
	clients  map[*client]struct{}
	upgrader websocket.Upgrader
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, initial any) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &client{conn: conn, send: make(chan []byte, 16)}
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	if initial != nil {
		if payload, err := json.Marshal(initial); err == nil {
			c.send <- payload
		}
	}
	go h.writeLoop(c)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
	h.remove(c)
}

func (h *Hub) Broadcast(event any) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.send <- payload:
		default:
		}
	}
}

func (h *Hub) writeLoop(c *client) {
	defer c.conn.Close()
	for payload := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			return
		}
	}
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	if _, exists := h.clients[c]; exists {
		delete(h.clients, c)
		close(c.send)
	}
	h.mu.Unlock()
	_ = c.conn.Close()
}
