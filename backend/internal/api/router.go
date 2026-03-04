package api

import (
	"net/http"
	"path"

	"github.com/dkrizic/webrtc/backend/internal/config"
	"github.com/dkrizic/webrtc/backend/internal/signaling"
)

func NewRouter(cfg *config.Config) *http.ServeMux {
	base := path.Clean("/" + cfg.APIBasePath)
	mux := http.NewServeMux()
	mux.HandleFunc(base+"/health", HealthHandler)
	mux.HandleFunc(base+"/status", StatusHandler)
	mux.HandleFunc("/ws", signaling.Handler)
	return mux
}
