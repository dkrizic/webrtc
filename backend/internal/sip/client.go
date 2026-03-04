package sip

import (
	"context"
	"log/slog"

	"github.com/dkrizic/webrtc/backend/internal/config"
)

// Client is a placeholder SIP user agent.
type Client struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Client {
	return &Client{cfg: cfg}
}

// Register attempts to register with the SIP server (stub).
func (c *Client) Register(ctx context.Context) error {
	slog.InfoContext(ctx, "SIP registration stub", "server", c.cfg.SIPServer)
	// TODO: implement SIP registration
	return nil
}
