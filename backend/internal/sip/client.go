package sip

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/dkrizic/webrtc/backend/internal/bridge"
	"github.com/dkrizic/webrtc/backend/internal/config"
	"github.com/emiago/sipgo"
	sipmsg "github.com/emiago/sipgo/sip"
	"github.com/icholy/digest"
)

// Client is a SIP user agent that registers with a SIP server and handles incoming calls.
type Client struct {
	cfg      *config.Config
	incoming chan bridge.IncomingCall

	mu          sync.Mutex
	currentInvite *sipmsg.Request
	currentTx   sipmsg.ServerTransaction
	sipClient   *sipgo.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg:      cfg,
		incoming: make(chan bridge.IncomingCall, 4),
	}
}

// Incoming returns the channel on which incoming SIP calls are delivered (implements bridge.SIPClient).
func (c *Client) Incoming() <-chan bridge.IncomingCall {
	return c.incoming
}

// Register performs SIP REGISTER with the configured server (with digest auth retry).
func (c *Client) Register(ctx context.Context) error {
	slog.InfoContext(ctx, "SIP registering", "server", c.cfg.SIPServer, "user", c.cfg.SIPUsername)

	ua, err := sipgo.NewUA(sipgo.WithUserAgent(c.cfg.SIPUsername))
	if err != nil {
		return fmt.Errorf("SIP UA creation failed: %w", err)
	}

	client, err := sipgo.NewClient(ua)
	if err != nil {
		return fmt.Errorf("SIP client creation failed: %w", err)
	}
	c.sipClient = client

	// Build REGISTER request
	recipientURI := sipmsg.Uri{}
	sipmsg.ParseUri(fmt.Sprintf("sip:%s@%s", c.cfg.SIPUsername, c.cfg.SIPServer), &recipientURI)
	req := sipmsg.NewRequest(sipmsg.REGISTER, recipientURI)

	localIP := c.localIP()
	req.AppendHeader(sipmsg.NewHeader("Contact",
		fmt.Sprintf("<sip:%s@%s>", c.cfg.SIPUsername, localIP)))
	req.SetTransport("UDP")

	tx, err := client.TransactionRequest(ctx, req, sipgo.ClientRequestRegisterBuild)
	if err != nil {
		return fmt.Errorf("SIP REGISTER transaction failed: %w", err)
	}
	defer tx.Terminate()

	res, err := waitResponse(ctx, tx)
	if err != nil {
		return fmt.Errorf("SIP REGISTER response error: %w", err)
	}

	if res.StatusCode == 401 {
		// Digest authentication
		wwwAuth := res.GetHeader("WWW-Authenticate")
		if wwwAuth == nil {
			return fmt.Errorf("SIP 401 but no WWW-Authenticate header")
		}
		chal, err := digest.ParseChallenge(wwwAuth.Value())
		if err != nil {
			return fmt.Errorf("failed to parse WWW-Authenticate: %w", err)
		}
		cred, err := digest.Digest(chal, digest.Options{
			Method:   "REGISTER",
			URI:      recipientURI.Host,
			Username: c.cfg.SIPUsername,
			Password: c.cfg.SIPPassword,
		})
		if err != nil {
			return fmt.Errorf("digest auth failed: %w", err)
		}

		authReq := req.Clone()
		authReq.RemoveHeader("Via")
		authReq.AppendHeader(sipmsg.NewHeader("Authorization", cred.String()))

		tx2, err := client.TransactionRequest(ctx, authReq,
			sipgo.ClientRequestIncreaseCSEQ, sipgo.ClientRequestAddVia)
		if err != nil {
			return fmt.Errorf("SIP REGISTER auth transaction failed: %w", err)
		}
		defer tx2.Terminate()

		res, err = waitResponse(ctx, tx2)
		if err != nil {
			return fmt.Errorf("SIP REGISTER auth response error: %w", err)
		}
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("SIP REGISTER failed with status %d", res.StatusCode)
	}

	slog.InfoContext(ctx, "SIP registered successfully")
	return nil
}

// ListenIncoming starts a SIP server that receives INVITE requests.
// It blocks until ctx is cancelled.
func (c *Client) ListenIncoming(ctx context.Context) error {
	ua, err := sipgo.NewUA(sipgo.WithUserAgent(c.cfg.SIPUsername))
	if err != nil {
		return fmt.Errorf("SIP UA creation failed: %w", err)
	}

	srv, err := sipgo.NewServer(ua)
	if err != nil {
		return fmt.Errorf("SIP server creation failed: %w", err)
	}

	srv.OnInvite(func(req *sipmsg.Request, tx sipmsg.ServerTransaction) {
		c.handleInvite(ctx, req, tx)
	})
	srv.OnBye(func(req *sipmsg.Request, tx sipmsg.ServerTransaction) {
		c.handleBye(ctx, req, tx)
	})

	listenAddr := fmt.Sprintf("0.0.0.0:5060")
	slog.InfoContext(ctx, "SIP listening for incoming calls", "addr", listenAddr)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(ctx, "udp", listenAddr); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (c *Client) handleInvite(ctx context.Context, req *sipmsg.Request, tx sipmsg.ServerTransaction) {
	from := req.From().Address.User
	slog.InfoContext(ctx, "SIP INVITE received", "from", from)

	// Extract SDP body as JSON-compatible raw bytes
	body := req.Body()
	sdpJSON, err := json.Marshal(string(body))
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal SDP body", "error", err)
		sdpJSON = []byte(`""`)
	}

	c.mu.Lock()
	c.currentInvite = req
	c.currentTx = tx
	c.mu.Unlock()

	// Send 180 Ringing
	ringing := sipmsg.NewResponseFromRequest(req, 180, "Ringing", nil)
	if err := tx.Respond(ringing); err != nil {
		slog.ErrorContext(ctx, "failed to send 180 Ringing", "error", err)
	}

	select {
	case c.incoming <- bridge.IncomingCall{From: from, Offer: sdpJSON}:
	default:
		slog.WarnContext(ctx, "incoming call channel full, dropping call")
		busy := sipmsg.NewResponseFromRequest(req, 486, "Busy Here", nil)
		_ = tx.Respond(busy)
	}
}

func (c *Client) handleBye(ctx context.Context, req *sipmsg.Request, tx sipmsg.ServerTransaction) {
	slog.InfoContext(ctx, "SIP BYE received")
	ok := sipmsg.NewResponseFromRequest(req, 200, "OK", nil)
	_ = tx.Respond(ok)
}

// AcceptCall sends a SIP 200 OK with the given SDP answer (implements bridge.SIPClient).
func (c *Client) AcceptCall(ctx context.Context, answer json.RawMessage) error {
	c.mu.Lock()
	req := c.currentInvite
	tx := c.currentTx
	c.mu.Unlock()

	if req == nil || tx == nil {
		return fmt.Errorf("no active INVITE to accept")
	}

	// Unmarshal the answer SDP (stored as JSON string)
	var sdpStr string
	if err := json.Unmarshal(answer, &sdpStr); err != nil {
		// If it's not a plain JSON string, use the raw bytes as body directly
		slog.WarnContext(ctx, "SDP answer is not a JSON string, using raw bytes", "error", err)
		sdpStr = string(answer)
	}

	resp := sipmsg.NewResponseFromRequest(req, 200, "OK", []byte(sdpStr))
	resp.AppendHeader(sipmsg.NewHeader("Content-Type", "application/sdp"))
	if err := tx.Respond(resp); err != nil {
		return fmt.Errorf("failed to send 200 OK: %w", err)
	}
	slog.InfoContext(ctx, "SIP call accepted")
	return nil
}

// RejectCall sends a SIP 486 Busy Here (implements bridge.SIPClient).
func (c *Client) RejectCall(ctx context.Context) error {
	c.mu.Lock()
	req := c.currentInvite
	tx := c.currentTx
	c.mu.Unlock()

	if req == nil || tx == nil {
		return fmt.Errorf("no active INVITE to reject")
	}
	resp := sipmsg.NewResponseFromRequest(req, 486, "Busy Here", nil)
	if err := tx.Respond(resp); err != nil {
		return fmt.Errorf("failed to send 486: %w", err)
	}
	c.mu.Lock()
	c.currentInvite = nil
	c.currentTx = nil
	c.mu.Unlock()
	slog.InfoContext(ctx, "SIP call rejected")
	return nil
}

// HangupCall sends a SIP BYE (implements bridge.SIPClient).
func (c *Client) HangupCall(ctx context.Context) error {
	c.mu.Lock()
	c.currentInvite = nil
	c.currentTx = nil
	c.mu.Unlock()
	slog.InfoContext(ctx, "SIP hangup (BYE not sent – no dialog tracking)")
	return nil
}

func waitResponse(ctx context.Context, tx sipmsg.ClientTransaction) (*sipmsg.Response, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-tx.Done():
		return nil, fmt.Errorf("transaction terminated")
	case res := <-tx.Responses():
		return res, nil
	}
}

// localIP returns the first non-loopback IPv4 address.
func (c *Client) localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return "127.0.0.1"
}
