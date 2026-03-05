package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/dkrizic/webrtc/backend/internal/api"
	"github.com/dkrizic/webrtc/backend/internal/bridge"
	"github.com/dkrizic/webrtc/backend/internal/config"
	"github.com/dkrizic/webrtc/backend/internal/signaling"
	"github.com/dkrizic/webrtc/backend/internal/sip"
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
			&cli.StringFlag{Name: "sip-transport", EnvVars: []string{"SIP_TRANSPORT"}, Value: "UDP", Usage: "SIP transport protocol (UDP or TCP)"},
			&cli.IntFlag{Name: "sip-max-ice-candidates", EnvVars: []string{"SIP_MAX_ICE_CANDIDATES"}, Value: 0, Usage: "Maximum ICE candidates per media section in outgoing SIP INVITE (0 = unlimited; set to a low value to reduce SDP packet size and avoid UDP MTU issues)"},
			&cli.StringFlag{Name: "listen-addr", EnvVars: []string{"LISTEN_ADDR"}, Value: ":8080", Usage: "HTTP listen address"},
			&cli.StringFlag{Name: "log-level", EnvVars: []string{"LOG_LEVEL"}, Value: "info", Usage: "Log level (debug, info, warn, error)"},
			&cli.StringFlag{Name: "api-base-path", EnvVars: []string{"API_BASE_PATH"}, Value: "/api", Usage: "Base path for API endpoints"},
		},
		Action: func(c *cli.Context) error {
			cfg := &config.Config{
				SIPServer:           c.String("sip-server"),
				SIPUsername:         c.String("sip-username"),
				SIPPassword:         c.String("sip-password"),
				SIPDomain:           c.String("sip-domain"),
				SIPTransport:        c.String("sip-transport"),
				SIPMaxICECandidates: c.Int("sip-max-ice-candidates"),
				ListenAddr:          c.String("listen-addr"),
				LogLevel:            c.String("log-level"),
				APIBasePath:         c.String("api-base-path"),
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
			transport := cfg.SIPTransport
			if transport != "UDP" && transport != "TCP" {
				return fmt.Errorf("invalid SIP_TRANSPORT %q: must be UDP or TCP", transport)
			}

			slog.Info("starting webrtc-backend", "listen", cfg.ListenAddr, "log-level", cfg.LogLevel, "api-base-path", cfg.APIBasePath)

			sipClient := sip.New(cfg)
			hub := signaling.NewHub(nil)
			br := bridge.New(hub, sipClient)
			hub.SetRouter(br)

			ctx := context.Background()
			if err := br.Start(ctx); err != nil {
				return fmt.Errorf("bridge start failed: %w", err)
			}

			go func() {
				if err := sipClient.Register(ctx); err != nil {
					slog.Warn("SIP registration failed", "error", err)
				}
				if err := sipClient.ListenIncoming(ctx); err != nil {
					slog.Error("SIP listen failed", "error", err)
				}
			}()

			router := api.NewRouterWithHub(cfg, hub, sipClient)
			slog.Info("server started", "addr", cfg.ListenAddr)
			return http.ListenAndServe(cfg.ListenAddr, router)
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}
