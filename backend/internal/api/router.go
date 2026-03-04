package api

import (
	"net/http"

	"github.com/dkrizic/webrtc/backend/internal/signaling"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", HealthHandler)
	mux.HandleFunc("/api/status", StatusHandler)
	mux.HandleFunc("/ws", signaling.Handler)
	return mux
}
