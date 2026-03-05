package bridge

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dkrizic/webrtc/backend/internal/signaling"
)

// fakeSIP is a test double for the SIPClient interface.
type fakeSIP struct {
	incoming    chan IncomingCall
	acceptCalls []json.RawMessage
	hangupCalls int
	rejectCalls int
	makeCalls   []struct {
		to    string
		offer json.RawMessage
	}
}

func newFakeSIP() *fakeSIP {
	return &fakeSIP{incoming: make(chan IncomingCall, 4)}
}

func (f *fakeSIP) AcceptCall(_ context.Context, answer json.RawMessage) error {
	f.acceptCalls = append(f.acceptCalls, answer)
	return nil
}
func (f *fakeSIP) RejectCall(_ context.Context) error { f.rejectCalls++; return nil }
func (f *fakeSIP) HangupCall(_ context.Context) error { f.hangupCalls++; return nil }
func (f *fakeSIP) MakeCall(_ context.Context, to string, offer json.RawMessage) error {
	f.makeCalls = append(f.makeCalls, struct {
		to    string
		offer json.RawMessage
	}{to: to, offer: offer})
	return nil
}
func (f *fakeSIP) Incoming() <-chan IncomingCall { return f.incoming }

func TestBridge_ForwardsIncomingToHub(t *testing.T) {
	hub := signaling.NewHub(nil)
	sip := newFakeSIP()
	br := New(hub, sip)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := br.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Simulate an incoming SIP call
	sip.incoming <- IncomingCall{From: "bob", Offer: json.RawMessage(`"v=0\r\n"`)}

	// Give the forwardIncoming goroutine time to run
	time.Sleep(50 * time.Millisecond)
	// No connected clients — just ensure no panic/deadlock occurred.
}

func TestBridge_RouteAnswer(t *testing.T) {
	hub := signaling.NewHub(nil)
	sip := newFakeSIP()
	br := New(hub, sip)

	ctx := context.Background()
	br.Start(ctx) //nolint:errcheck

	var sent []signaling.Message
	sendFn := func(msg signaling.Message) error {
		sent = append(sent, msg)
		return nil
	}

	br.Route(signaling.Message{Type: signaling.TypeAnswer, Payload: map[string]interface{}{"sdp": "answer"}}, sendFn)

	if len(sip.acceptCalls) != 1 {
		t.Errorf("expected 1 AcceptCall, got %d", len(sip.acceptCalls))
	}
}

func TestBridge_RouteHangup(t *testing.T) {
	hub := signaling.NewHub(nil)
	sip := newFakeSIP()
	br := New(hub, sip)

	ctx := context.Background()
	br.Start(ctx) //nolint:errcheck

	br.Route(signaling.Message{Type: signaling.TypeHangup}, func(signaling.Message) error { return nil })

	if sip.hangupCalls != 1 {
		t.Errorf("expected 1 HangupCall, got %d", sip.hangupCalls)
	}
}

func TestBridge_RouteDial(t *testing.T) {
	hub := signaling.NewHub(nil)
	sip := newFakeSIP()
	br := New(hub, sip)

	ctx := context.Background()
	br.Start(ctx) //nolint:errcheck

	br.Route(signaling.Message{
		Type:    signaling.TypeDial,
		Payload: map[string]interface{}{"to": "1234567890"},
	}, func(signaling.Message) error { return nil })

	br.mu.Lock()
	pendingTo := br.pendingTo
	br.mu.Unlock()

	if pendingTo != "1234567890" {
		t.Errorf("expected pendingTo %q, got %q", "1234567890", pendingTo)
	}
}

func TestBridge_RouteDialData(t *testing.T) {
	hub := signaling.NewHub(nil)
	sip := newFakeSIP()
	br := New(hub, sip)

	ctx := context.Background()
	br.Start(ctx) //nolint:errcheck

	// TypeDial with 'to' carried in msg.Data instead of payload.
	br.Route(signaling.Message{
		Type: signaling.TypeDial,
		Data: "9876543210",
	}, func(signaling.Message) error { return nil })

	br.mu.Lock()
	pendingTo := br.pendingTo
	br.mu.Unlock()

	if pendingTo != "9876543210" {
		t.Errorf("expected pendingTo %q, got %q", "9876543210", pendingTo)
	}
}

func TestBridge_RouteOffer_MakesCall(t *testing.T) {
	hub := signaling.NewHub(nil)
	sip := newFakeSIP()
	br := New(hub, sip)

	ctx := context.Background()
	br.Start(ctx) //nolint:errcheck

	// First store the 'to' via TypeDial.
	br.Route(signaling.Message{
		Type:    signaling.TypeDial,
		Payload: map[string]interface{}{"to": "alice"},
	}, func(signaling.Message) error { return nil })

	// Then send the offer.
	br.Route(signaling.Message{
		Type:    signaling.TypeOffer,
		Payload: map[string]interface{}{"sdp": "v=0\r\n"},
	}, func(signaling.Message) error { return nil })

	if len(sip.makeCalls) != 1 {
		t.Fatalf("expected 1 MakeCall, got %d", len(sip.makeCalls))
	}
	if sip.makeCalls[0].to != "alice" {
		t.Errorf("expected MakeCall 'to' %q, got %q", "alice", sip.makeCalls[0].to)
	}
}

func TestBridge_RouteOffer_NoPendingTo(t *testing.T) {
	hub := signaling.NewHub(nil)
	sip := newFakeSIP()
	br := New(hub, sip)

	ctx := context.Background()
	br.Start(ctx) //nolint:errcheck

	// Send offer without prior dial — MakeCall must not be invoked.
	br.Route(signaling.Message{
		Type:    signaling.TypeOffer,
		Payload: map[string]interface{}{"sdp": "v=0\r\n"},
	}, func(signaling.Message) error { return nil })

	if len(sip.makeCalls) != 0 {
		t.Errorf("expected 0 MakeCall, got %d", len(sip.makeCalls))
	}
}
