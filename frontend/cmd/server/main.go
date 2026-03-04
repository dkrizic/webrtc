package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

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

			// Check backend connectivity before starting
			healthURL := backendURL + "/api/health"
			client := &http.Client{Timeout: 5 * time.Second}
			var lastErr error
			for i := 0; i < 5; i++ {
				resp, err := client.Get(healthURL)
				if err == nil && resp.StatusCode == 200 {
					resp.Body.Close()
					lastErr = nil
					slog.Info("backend health check passed", "url", healthURL)
					break
				}
				if err != nil {
					lastErr = err
				} else {
					resp.Body.Close()
					lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				}
				slog.Warn("backend health check failed, retrying...", "attempt", i+1, "error", lastErr)
				time.Sleep(2 * time.Second)
			}
			if lastErr != nil {
				return fmt.Errorf("backend is not accessible at %s: %w", healthURL, lastErr)
			}

			// Serve static files
			fs := http.FileServer(http.Dir("/app/static"))
			mux := http.NewServeMux()
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
