package signaling

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.ErrorContext(ctx, "websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()
	slog.InfoContext(ctx, "websocket client connected", "remote", r.RemoteAddr)
	// Send initial status
	msg := Message{Type: TypeStatus, Data: "connected"}
	if err := conn.WriteJSON(msg); err != nil {
		slog.ErrorContext(ctx, "failed to send status", "error", err)
		return
	}
	// Read loop (stub)
	for {
		var incoming Message
		if err := conn.ReadJSON(&incoming); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.ErrorContext(ctx, "websocket read error", "error", err)
			}
			break
		}
		slog.InfoContext(ctx, "received message", "type", incoming.Type)
		// TODO: route to bridge
	}
	slog.InfoContext(ctx, "websocket client disconnected", "remote", r.RemoteAddr)
}
