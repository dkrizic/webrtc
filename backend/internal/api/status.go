package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		slog.ErrorContext(r.Context(), "health encode error", "error", err)
	}
	slog.InfoContext(r.Context(), "health check")
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"sip": "unregistered"}); err != nil {
		slog.ErrorContext(r.Context(), "status encode error", "error", err)
	}
	slog.InfoContext(r.Context(), "status check")
}
