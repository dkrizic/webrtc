package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dkrizic/webrtc/backend/internal/version"
)

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"version": version.Version}); err != nil {
		slog.ErrorContext(r.Context(), "version encode error", "error", err)
	}
	slog.DebugContext(r.Context(), "version check", "version", version.Version)
}
