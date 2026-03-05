package bridge

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

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
	// MakeCall initiates an outgoing SIP INVITE to 'to' with the given SDP offer.
	MakeCall(ctx context.Context, to string, offer json.RawMessage) error
	// Incoming returns a channel on which incoming SIP calls are delivered.
	Incoming() <-chan IncomingCall
}

// Bridge orchestrates between SIP incoming calls and WebSocket-connected clients.
type Bridge struct {
	hub       *signaling.Hub
	sip       SIPClient
	ctx       context.Context
	mu        sync.Mutex
	pendingTo string
}

// New creates a Bridge. hub must not be nil. sipClient may be nil (no SIP support).
func New(hub *signaling.Hub, sipClient SIPClient) *Bridge {
	return &Bridge{hub: hub, sip: sipClient, ctx: context.Background()}
}

// Start initializes the bridge and begins forwarding SIP events to WebSocket clients.
func (b *Bridge) Start(ctx context.Context) error {
	b.ctx = ctx
	slog.InfoContext(ctx, "Bridge: initialized and starting")
	if b.sip != nil {
		slog.DebugContext(ctx, "Bridge: SIP client available, starting incoming call forwarder")
		go b.forwardIncoming(ctx)
	} else {
		slog.WarnContext(ctx, "Bridge: no SIP client configured")
	}
	return nil
}

// forwardIncoming listens for incoming SIP calls and notifies WebSocket clients.
func (b *Bridge) forwardIncoming(ctx context.Context) {
	ch := b.sip.Incoming()
	slog.DebugContext(ctx, "Bridge: waiting for incoming SIP calls...")
	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Bridge: incoming call forwarder shutting down")
			return
		case call, ok := <-ch:
			if !ok {
				slog.WarnContext(ctx, "Bridge: incoming call channel closed")
				return
			}
			slog.InfoContext(ctx, "Bridge: 🔔 forwarding incoming call to WebSocket clients", "from", call.From)
			b.hub.Broadcast(signaling.Message{
				Type: signaling.TypeIncoming,
				Payload: map[string]interface{}{
					"from":  call.From,
					"offer": call.Offer,
				},
			})
			slog.DebugContext(ctx, "Bridge: incoming call broadcasted", "from", call.From)
		}
	}
}

// Route implements signaling.Router — called for every message the WebSocket client sends.
func (b *Bridge) Route(msg signaling.Message, send func(signaling.Message) error) {
	ctx := b.ctx
	switch msg.Type {
	case signaling.TypeDial:
		if b.sip == nil {
			slog.WarnContext(ctx, "Bridge: received dial but no SIP client configured")
			return
		}
		to := extractTo(msg)
		if to == "" {
			slog.WarnContext(ctx, "Bridge: received dial but 'to' field is empty")
			return
		}
		b.mu.Lock()
		b.pendingTo = to
		b.mu.Unlock()
		slog.InfoContext(ctx, "Bridge: 📞 received dial, stored 'to' number", "to", to)
	case signaling.TypeOffer:
		if b.sip == nil {
			slog.WarnContext(ctx, "Bridge: received offer but no SIP client configured")
			return
		}
		b.mu.Lock()
		to := b.pendingTo
		b.mu.Unlock()
		if to == "" {
			slog.WarnContext(ctx, "Bridge: received offer but no pending 'to' number; ignoring")
			return
		}
		raw, err := json.Marshal(msg.Payload)
		if err != nil {
			slog.ErrorContext(ctx, "Bridge: failed to marshal offer", "error", err)
			return
		}
		slog.InfoContext(ctx, "Bridge: 📤 received offer, initiating outgoing SIP call", "to", to)
		if err := b.sip.MakeCall(ctx, to, raw); err != nil {
			slog.ErrorContext(ctx, "Bridge: MakeCall failed", "error", err)
			_ = send(signaling.Message{Type: signaling.TypeError, Data: err.Error()})
			return
		}
	case signaling.TypeAnswer:
		if b.sip == nil {
			slog.WarnContext(ctx, "Bridge: received answer but no SIP client configured")
			return
		}
		slog.InfoContext(ctx, "Bridge: ✅ received answer from WebSocket, accepting SIP call")
		raw, err := json.Marshal(msg.Payload)
		if err != nil {
			slog.ErrorContext(ctx, "Bridge: failed to marshal answer", "error", err)
			return
		}
		if err := b.sip.AcceptCall(ctx, raw); err != nil {
			slog.ErrorContext(ctx, "Bridge: AcceptCall failed", "error", err)
		}
	case signaling.TypeHangup:
		if b.sip == nil {
			slog.WarnContext(ctx, "Bridge: received hangup but no SIP client configured")
			return
		}
		slog.InfoContext(ctx, "Bridge: ❌ received hangup from WebSocket, ending SIP call")
		if err := b.sip.HangupCall(ctx); err != nil {
			slog.ErrorContext(ctx, "Bridge: HangupCall failed", "error", err)
		}
	default:
		slog.DebugContext(ctx, "Bridge: unhandled message type", "type", msg.Type)
	}
}

// extractTo retrieves the 'to' destination from a dial message.
// It checks msg.Payload (as a map with a "to" key) and falls back to msg.Data.
func extractTo(msg signaling.Message) string {
	if payload, ok := msg.Payload.(map[string]interface{}); ok {
		if v, ok := payload["to"].(string); ok && v != "" {
			return v
		}
	}
	return msg.Data
}
