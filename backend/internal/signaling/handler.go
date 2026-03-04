package signaling

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Router receives messages from WebSocket clients and forwards them for processing.
type Router interface {
	Route(msg Message, send func(Message) error)
}

// Hub manages connected WebSocket clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[*wsClient]struct{}
	router  Router
}

// clientSendBuffer is the number of outbound messages buffered per WebSocket client.
// A buffer of 32 prevents blocking the read loop on slow writers without using
// unbounded memory.
const clientSendBuffer = 32

// wsClient represents a single connected WebSocket peer.
type wsClient struct {
	conn *websocket.Conn
	send chan Message
}

// NewHub creates a new Hub with the given message router.
func NewHub(router Router) *Hub {
	return &Hub{
		clients: make(map[*wsClient]struct{}),
		router:  router,
	}
}

// SetRouter updates the message router (safe to call before any clients connect).
func (h *Hub) SetRouter(r Router) {
	h.mu.Lock()
	h.router = r
	h.mu.Unlock()
}

// Broadcast sends a message to all connected WebSocket clients.
func (h *Hub) Broadcast(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
		}
	}
}

func (h *Hub) register(c *wsClient) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(c *wsClient) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

// ServeWS upgrades an HTTP connection to WebSocket and handles the client lifecycle.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.ErrorContext(ctx, "websocket upgrade failed", "error", err)
		return
	}
	c := &wsClient{conn: conn, send: make(chan Message, clientSendBuffer)}
	h.register(c)
	defer func() {
		h.unregister(c)
		conn.Close()
		slog.InfoContext(ctx, "websocket client disconnected", "remote", r.RemoteAddr)
	}()
	slog.InfoContext(ctx, "websocket client connected", "remote", r.RemoteAddr)

	// Send initial status
	initial := Message{Type: TypeStatus, Data: "connected"}
	if err := conn.WriteJSON(initial); err != nil {
		slog.ErrorContext(ctx, "failed to send status", "error", err)
		return
	}

	// Writer goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		for msg := range c.send {
			if err := conn.WriteJSON(msg); err != nil {
				slog.ErrorContext(ctx, "websocket write error", "error", err)
				return
			}
		}
	}()

	sendFn := func(msg Message) error {
		select {
		case c.send <- msg:
		default:
		}
		return nil
	}

	// Read loop
	for {
		var incoming Message
		if err := conn.ReadJSON(&incoming); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.ErrorContext(ctx, "websocket read error", "error", err)
			}
			break
		}
		slog.InfoContext(ctx, "received message", "type", incoming.Type)
		if h.router != nil {
			h.router.Route(incoming, sendFn)
		}
	}
	close(c.send)
	<-done
}

// Handler is a convenience http.HandlerFunc that uses a package-level default Hub (no routing).
func Handler(w http.ResponseWriter, r *http.Request) {
	defaultHub.ServeWS(w, r)
}

// DefaultHub returns the package-level default Hub instance.
func DefaultHub() *Hub {
	return defaultHub
}

var defaultHub = NewHub(nil)
