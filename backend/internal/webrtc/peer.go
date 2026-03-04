package webrtc

import (
	"context"
	"log/slog"

	pionwebrtc "github.com/pion/webrtc/v3"
)

// PeerConnection manages a single WebRTC peer connection using pion/webrtc.
type PeerConnection struct {
	pc *pionwebrtc.PeerConnection
}

func New() *PeerConnection {
	return &PeerConnection{}
}

// Init creates the underlying pion PeerConnection with a default STUN server.
func (p *PeerConnection) Init(ctx context.Context) error {
	cfg := pionwebrtc.Configuration{
		ICEServers: []pionwebrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}
	pc, err := pionwebrtc.NewPeerConnection(cfg)
	if err != nil {
		return err
	}
	p.pc = pc

	pc.OnICEConnectionStateChange(func(state pionwebrtc.ICEConnectionState) {
		slog.InfoContext(ctx, "WebRTC ICE state changed", "state", state.String())
	})

	slog.InfoContext(ctx, "WebRTC peer connection initialized")
	return nil
}

// SetRemoteOffer sets the remote SDP offer and returns the local SDP answer.
func (p *PeerConnection) SetRemoteOffer(ctx context.Context, offerSDP string) (string, error) {
	offer := pionwebrtc.SessionDescription{
		Type: pionwebrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}
	if err := p.pc.SetRemoteDescription(offer); err != nil {
		return "", err
	}
	answer, err := p.pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	if err := p.pc.SetLocalDescription(answer); err != nil {
		return "", err
	}
	slog.InfoContext(ctx, "WebRTC answer created")
	return answer.SDP, nil
}

// Close tears down the peer connection.
func (p *PeerConnection) Close() error {
	if p.pc != nil {
		return p.pc.Close()
	}
	return nil
}

// OnICECandidate registers a callback for local ICE candidates.
func (p *PeerConnection) OnICECandidate(fn func(*pionwebrtc.ICECandidate)) {
	if p.pc != nil {
		p.pc.OnICECandidate(fn)
	}
}

// AddICECandidate adds a remote ICE candidate to the peer connection.
func (p *PeerConnection) AddICECandidate(candidate pionwebrtc.ICECandidateInit) error {
	if p.pc != nil {
		return p.pc.AddICECandidate(candidate)
	}
	return nil
}
