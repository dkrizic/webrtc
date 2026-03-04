package api

import (
	"net/http"
	"path"

	"github.com/dkrizic/webrtc/backend/internal/config"
	"github.com/dkrizic/webrtc/backend/internal/signaling"
)

func NewRouter(cfg *config.Config) *http.ServeMux {
	return NewRouterWithHub(cfg, signaling.DefaultHub())
}

func NewRouterWithHub(cfg *config.Config, hub *signaling.Hub) *http.ServeMux {
	base := path.Clean("/" + cfg.APIBasePath)
	mux := http.NewServeMux()
	mux.HandleFunc(base+"/health", HealthHandler)
	mux.HandleFunc(base+"/status", StatusHandler(cfg))
	mux.HandleFunc("/ws", hub.ServeWS)
	return mux
}
