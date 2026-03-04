package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dkrizic/webrtc/backend/internal/config"
)

// SIPStatusProvider provides the current SIP registration status
type SIPStatusProvider interface {
	IsRegistered() bool
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		slog.ErrorContext(r.Context(), "health encode error", "error", err)
	}
	slog.DebugContext(r.Context(), "health check")
}

func StatusHandler(cfg *config.Config, sipProvider SIPStatusProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		sipStatus := "unregistered"
		if sipProvider != nil && sipProvider.IsRegistered() {
			sipStatus = "registered"
		}

		if err := json.NewEncoder(w).Encode(map[string]string{
			"sip":          sipStatus,
			"phone_number": cfg.SIPUsername,
		}); err != nil {
			slog.ErrorContext(r.Context(), "status encode error", "error", err)
		}
		slog.DebugContext(r.Context(), "status check", "sip_status", sipStatus)
	}
}
