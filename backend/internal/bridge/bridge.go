package bridge

import (
	"context"
	"log/slog"
)

// Bridge orchestrates between WebRTC peer connections and SIP calls (stub).
type Bridge struct{}

func New() *Bridge {
	return &Bridge{}
}

// Start initializes the bridge (stub).
func (b *Bridge) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "bridge stub started")
	// TODO: implement bridge logic
	return nil
}
