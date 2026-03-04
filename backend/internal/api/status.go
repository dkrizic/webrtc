package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dkrizic/webrtc/backend/internal/config"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		slog.ErrorContext(r.Context(), "health encode error", "error", err)
	}
	slog.InfoContext(r.Context(), "health check")
}

func StatusHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"sip": "unregistered", "phone_number": cfg.SIPUsername}); err != nil {
			slog.ErrorContext(r.Context(), "status encode error", "error", err)
		}
		slog.InfoContext(r.Context(), "status check")
	}
}
