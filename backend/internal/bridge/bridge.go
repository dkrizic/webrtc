package bridge

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/dkrizic/webrtc/backend/internal/signaling"
)

// IncomingCall carries the data for a SIP INVITE received by the SIP client.
type IncomingCall struct {
	From  string
	Offer json.RawMessage
}

// SIPClient is the interface the bridge uses to interact with the SIP layer.
type SIPClient interface {
	// AcceptCall sends a SIP 200 OK with the given SDP answer.
	AcceptCall(ctx context.Context, answer json.RawMessage) error
	// RejectCall sends a SIP rejection response.
	RejectCall(ctx context.Context) error
	// HangupCall sends a SIP BYE.
	HangupCall(ctx context.Context) error
	// Incoming returns a channel on which incoming SIP calls are delivered.
	Incoming() <-chan IncomingCall
}

// Bridge orchestrates between SIP incoming calls and WebSocket-connected clients.
type Bridge struct {
	hub *signaling.Hub
	sip SIPClient
	ctx context.Context
}

// New creates a Bridge. hub must not be nil. sipClient may be nil (no SIP support).
func New(hub *signaling.Hub, sipClient SIPClient) *Bridge {
	return &Bridge{hub: hub, sip: sipClient, ctx: context.Background()}
}

// Start initializes the bridge and begins forwarding SIP events to WebSocket clients.
func (b *Bridge) Start(ctx context.Context) error {
	b.ctx = ctx
	slog.InfoContext(ctx, "bridge started")
	if b.sip != nil {
		go b.forwardIncoming(ctx)
	}
	return nil
}

// forwardIncoming listens for incoming SIP calls and notifies WebSocket clients.
func (b *Bridge) forwardIncoming(ctx context.Context) {
	ch := b.sip.Incoming()
	for {
		select {
		case <-ctx.Done():
			return
		case call, ok := <-ch:
			if !ok {
				return
			}
			slog.InfoContext(ctx, "incoming SIP call", "from", call.From)
			b.hub.Broadcast(signaling.Message{
				Type: signaling.TypeIncoming,
				Payload: map[string]interface{}{
					"from":  call.From,
					"offer": call.Offer,
				},
			})
		}
	}
}

// Route implements signaling.Router — called for every message the WebSocket client sends.
func (b *Bridge) Route(msg signaling.Message, send func(signaling.Message) error) {
	ctx := b.ctx
	switch msg.Type {
	case signaling.TypeAnswer:
		if b.sip == nil {
			return
		}
		raw, err := json.Marshal(msg.Payload)
		if err != nil {
			slog.ErrorContext(ctx, "bridge: failed to marshal answer", "error", err)
			return
		}
		if err := b.sip.AcceptCall(ctx, raw); err != nil {
			slog.ErrorContext(ctx, "bridge: AcceptCall failed", "error", err)
		}
	case signaling.TypeHangup:
		if b.sip == nil {
			return
		}
		if err := b.sip.HangupCall(ctx); err != nil {
			slog.ErrorContext(ctx, "bridge: HangupCall failed", "error", err)
		}
	default:
		slog.InfoContext(ctx, "bridge: unhandled message type", "type", msg.Type)
	}
}