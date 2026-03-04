package webrtc

import (
	"context"
	"log/slog"
)

// PeerConnection manages a single WebRTC peer connection (stub).
type PeerConnection struct{}

func New() *PeerConnection {
	return &PeerConnection{}
}

// Init initializes the peer connection (stub).
func (p *PeerConnection) Init(ctx context.Context) error {
	slog.InfoContext(ctx, "WebRTC peer connection stub initialized")
	// TODO: implement using pion/webrtc
	return nil
}
