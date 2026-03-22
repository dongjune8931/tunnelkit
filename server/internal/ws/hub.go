package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub는 WebSocket 클라이언트(CLI)들을 관리한다.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // subdomain → client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) Register(subdomain string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[subdomain] = c
}

func (h *Hub) Unregister(subdomain string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, subdomain)
}

func (h *Hub) Get(subdomain string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	c, ok := h.clients[subdomain]
	return c, ok
}

func (h *Hub) List() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	subs := make([]string, 0, len(h.clients))
	for s := range h.clients {
		subs = append(subs, s)
	}
	return subs
}

// Client는 연결된 CLI WebSocket 클라이언트를 나타낸다.
type Client struct {
	Subdomain string
	Conn      *websocket.Conn
	mu        sync.Mutex
}

func NewClient(subdomain string, conn *websocket.Conn) *Client {
	return &Client{Subdomain: subdomain, Conn: conn}
}

func (c *Client) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}
