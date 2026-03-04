package main

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "webrtc-frontend",
		Usage: "WebRTC SIP frontend static server",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "backend-url", EnvVars: []string{"BACKEND_URL"}, Value: "http://backend:8080", Usage: "Backend URL"},
			&cli.StringFlag{Name: "listen-addr", EnvVars: []string{"LISTEN_ADDR"}, Value: ":3000", Usage: "HTTP listen address"},
			&cli.StringFlag{Name: "log-level", EnvVars: []string{"LOG_LEVEL"}, Value: "info", Usage: "Log level"},
		},
		Action: func(c *cli.Context) error {
			listenAddr := c.String("listen-addr")
			backendURL := c.String("backend-url")
			logLevel := c.String("log-level")

			var level slog.Level
			switch logLevel {
			case "debug":
				level = slog.LevelDebug
			case "warn":
				level = slog.LevelWarn
			case "error":
				level = slog.LevelError
			default:
				level = slog.LevelInfo
			}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
			slog.SetDefault(logger)

			slog.Info("starting webrtc-frontend", "listen", listenAddr, "backend", backendURL)

			target, err := url.Parse(backendURL)
			if err != nil {
				return err
			}
			proxy := httputil.NewSingleHostReverseProxy(target)

			mux := http.NewServeMux()

			// Reverse proxy for API and WebSocket
			mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
				slog.InfoContext(r.Context(), "proxying request", "path", r.URL.Path)
				proxy.ServeHTTP(w, r)
			})
			mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
				slog.InfoContext(r.Context(), "proxying websocket", "path", r.URL.Path)
				proxy.ServeHTTP(w, r)
			})

			// Serve static files
			fs := http.FileServer(http.Dir("/app/static"))
			mux.Handle("/", fs)

			slog.Info("server started", "addr", listenAddr)
			return http.ListenAndServe(listenAddr, mux)
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}
