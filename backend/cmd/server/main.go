package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/dkrizic/webrtc/backend/internal/api"
	"github.com/dkrizic/webrtc/backend/internal/config"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "webrtc-backend",
		Usage: "WebRTC to SIP gateway backend",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "sip-server", EnvVars: []string{"SIP_SERVER"}, Usage: "SIP server address", Required: true},
			&cli.StringFlag{Name: "sip-username", EnvVars: []string{"SIP_USERNAME"}, Usage: "SIP username", Required: true},
			&cli.StringFlag{Name: "sip-password", EnvVars: []string{"SIP_PASSWORD"}, Usage: "SIP password", Required: true},
			&cli.StringFlag{Name: "sip-domain", EnvVars: []string{"SIP_DOMAIN"}, Usage: "SIP domain", Required: true},
			&cli.StringFlag{Name: "listen-addr", EnvVars: []string{"LISTEN_ADDR"}, Value: ":8080", Usage: "HTTP listen address"},
			&cli.StringFlag{Name: "log-level", EnvVars: []string{"LOG_LEVEL"}, Value: "info", Usage: "Log level (debug, info, warn, error)"},
			&cli.StringFlag{Name: "api-base-path", EnvVars: []string{"API_BASE_PATH"}, Value: "/api", Usage: "Base path for API endpoints"},
		},
		Action: func(c *cli.Context) error {
			cfg := &config.Config{
				SIPServer:   c.String("sip-server"),
				SIPUsername: c.String("sip-username"),
				SIPPassword: c.String("sip-password"),
				SIPDomain:   c.String("sip-domain"),
				ListenAddr:  c.String("listen-addr"),
				LogLevel:    c.String("log-level"),
				APIBasePath: c.String("api-base-path"),
			}

			// Set up slog
			var level slog.Level
			switch cfg.LogLevel {
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

			// Validate required SIP configuration
			if cfg.SIPServer == "" {
				return fmt.Errorf("SIP_SERVER is required but not configured")
			}
			if cfg.SIPUsername == "" {
				return fmt.Errorf("SIP_USERNAME is required but not configured")
			}
			if cfg.SIPPassword == "" {
				return fmt.Errorf("SIP_PASSWORD is required but not configured")
			}
			if cfg.SIPDomain == "" {
				return fmt.Errorf("SIP_DOMAIN is required but not configured")
			}

			slog.Info("starting webrtc-backend", "listen", cfg.ListenAddr, "log-level", cfg.LogLevel, "api-base-path", cfg.APIBasePath)

			router := api.NewRouter(cfg)
			slog.Info("server started", "addr", cfg.ListenAddr)
			return http.ListenAndServe(cfg.ListenAddr, router)
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}
