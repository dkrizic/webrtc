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
func (f *fakeSIP) Incoming() <-chan IncomingCall      { return f.incoming }

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
