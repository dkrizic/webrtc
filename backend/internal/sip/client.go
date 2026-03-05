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

	mu            sync.Mutex
	registered    bool
	currentInvite *sipmsg.Request
	currentTx     sipmsg.ServerTransaction
	sipClient     *sipgo.Client
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

// IsRegistered returns whether the SIP client is currently registered.
func (c *Client) IsRegistered() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.registered
}

// Register performs SIP REGISTER with the configured server (with digest auth retry).
func (c *Client) Register(ctx context.Context) error {
	slog.InfoContext(ctx, "SIP: starting registration", "server", c.cfg.SIPServer, "username", c.cfg.SIPUsername, "domain", c.cfg.SIPDomain)

	ua, err := sipgo.NewUA(sipgo.WithUserAgent(c.cfg.SIPUsername))
	if err != nil {
		slog.ErrorContext(ctx, "SIP: failed to create user agent", "error", err)
		return fmt.Errorf("SIP UA creation failed: %w", err)
	}
	slog.DebugContext(ctx, "SIP: user agent created")

	client, err := sipgo.NewClient(ua)
	if err != nil {
		slog.ErrorContext(ctx, "SIP: failed to create client", "error", err)
		return fmt.Errorf("SIP client creation failed: %w", err)
	}
	c.sipClient = client
	slog.DebugContext(ctx, "SIP: client created")

	// Build REGISTER request
	recipientURI := sipmsg.Uri{}
	sipmsg.ParseUri(fmt.Sprintf("sip:%s@%s", c.cfg.SIPUsername, c.cfg.SIPServer), &recipientURI)
	req := sipmsg.NewRequest(sipmsg.REGISTER, recipientURI)
	slog.DebugContext(ctx, "SIP: REGISTER request built", "uri", recipientURI.String())

	localIP := c.localIP()
	req.AppendHeader(sipmsg.NewHeader("Contact",
		fmt.Sprintf("<sip:%s@%s>", c.cfg.SIPUsername, localIP)))
	req.SetTransport("UDP")
	slog.DebugContext(ctx, "SIP: contact header set", "local_ip", localIP)

	tx, err := client.TransactionRequest(ctx, req, sipgo.ClientRequestRegisterBuild)
	if err != nil {
		slog.ErrorContext(ctx, "SIP: REGISTER transaction failed", "error", err)
		return fmt.Errorf("SIP REGISTER transaction failed: %w", err)
	}
	defer tx.Terminate()

	res, err := waitResponse(ctx, tx)
	if err != nil {
		slog.ErrorContext(ctx, "SIP: no response to initial REGISTER", "error", err)
		return fmt.Errorf("SIP REGISTER response error: %w", err)
	}
	slog.DebugContext(ctx, "SIP: initial REGISTER response", "status_code", res.StatusCode)

	if res.StatusCode == 401 {
		slog.InfoContext(ctx, "SIP: received 401 Unauthorized, attempting digest authentication")
		// Digest authentication
		wwwAuth := res.GetHeader("WWW-Authenticate")
		if wwwAuth == nil {
			slog.ErrorContext(ctx, "SIP: 401 response but no WWW-Authenticate header")
			return fmt.Errorf("SIP 401 but no WWW-Authenticate header")
		}
		chal, err := digest.ParseChallenge(wwwAuth.Value())
		if err != nil {
			slog.ErrorContext(ctx, "SIP: failed to parse WWW-Authenticate", "error", err)
			return fmt.Errorf("failed to parse WWW-Authenticate: %w", err)
		}
		slog.DebugContext(ctx, "SIP: digest challenge parsed", "realm", chal.Realm)

		cred, err := digest.Digest(chal, digest.Options{
			Method:   "REGISTER",
			URI:      recipientURI.Host,
			Username: c.cfg.SIPUsername,
			Password: c.cfg.SIPPassword,
		})
		if err != nil {
			slog.ErrorContext(ctx, "SIP: digest authentication failed", "error", err)
			return fmt.Errorf("digest auth failed: %w", err)
		}
		slog.DebugContext(ctx, "SIP: digest credentials computed")

		authReq := req.Clone()
		authReq.RemoveHeader("Via")
		authReq.AppendHeader(sipmsg.NewHeader("Authorization", cred.String()))
		slog.DebugContext(ctx, "SIP: sending authenticated REGISTER request")

		tx2, err := client.TransactionRequest(ctx, authReq,
			sipgo.ClientRequestIncreaseCSEQ, sipgo.ClientRequestAddVia)
		if err != nil {
			slog.ErrorContext(ctx, "SIP: authenticated REGISTER transaction failed", "error", err)
			return fmt.Errorf("SIP REGISTER auth transaction failed: %w", err)
		}
		defer tx2.Terminate()

		res, err = waitResponse(ctx, tx2)
		if err != nil {
			slog.ErrorContext(ctx, "SIP: no response to authenticated REGISTER", "error", err)
			return fmt.Errorf("SIP REGISTER auth response error: %w", err)
		}
		slog.DebugContext(ctx, "SIP: authenticated REGISTER response", "status_code", res.StatusCode)
	}

	if res.StatusCode != 200 {
		slog.ErrorContext(ctx, "SIP: REGISTER failed with error response", "status_code", res.StatusCode, "reason", res.Reason)
		return fmt.Errorf("SIP REGISTER failed with status %d: %s", res.StatusCode, res.Reason)
	}

	c.mu.Lock()
	c.registered = true
	c.mu.Unlock()

	slog.InfoContext(ctx, "SIP: ✅ registration successful", "username", c.cfg.SIPUsername)
	return nil
}

// ListenIncoming starts a SIP server that receives INVITE requests.
// It blocks until ctx is cancelled.
func (c *Client) ListenIncoming(ctx context.Context) error {
	slog.InfoContext(ctx, "SIP: initializing incoming call listener")

	ua, err := sipgo.NewUA(sipgo.WithUserAgent(c.cfg.SIPUsername))
	if err != nil {
		slog.ErrorContext(ctx, "SIP: failed to create UA for listener", "error", err)
		return fmt.Errorf("SIP UA creation failed: %w", err)
	}
	slog.DebugContext(ctx, "SIP: listener UA created")

	srv, err := sipgo.NewServer(ua)
	if err != nil {
		slog.ErrorContext(ctx, "SIP: failed to create server", "error", err)
		return fmt.Errorf("SIP server creation failed: %w", err)
	}
	slog.DebugContext(ctx, "SIP: server created")

	srv.OnInvite(func(req *sipmsg.Request, tx sipmsg.ServerTransaction) {
		c.handleInvite(ctx, req, tx)
	})
	srv.OnBye(func(req *sipmsg.Request, tx sipmsg.ServerTransaction) {
		c.handleBye(ctx, req, tx)
	})

	listenAddr := fmt.Sprintf("0.0.0.0:5060")
	slog.InfoContext(ctx, "SIP: 📞 listener starting", "addr", listenAddr)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(ctx, "udp", listenAddr); err != nil {
			slog.ErrorContext(ctx, "SIP: listener error", "error", err)
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.InfoContext(ctx, "SIP: listener shutting down")
		return nil
	case err := <-errCh:
		return err
	}
}

func (c *Client) handleInvite(ctx context.Context, req *sipmsg.Request, tx sipmsg.ServerTransaction) {
	from := req.From().Address.User
	slog.InfoContext(ctx, "SIP: 🔔 incoming INVITE received", "from", from)

	// Extract SDP body as JSON-compatible raw bytes
	body := req.Body()
	sdpJSON, err := json.Marshal(string(body))
	if err != nil {
		slog.ErrorContext(ctx, "SIP: failed to marshal SDP body", "error", err)
		sdpJSON = []byte(`""`)
	}
	slog.DebugContext(ctx, "SIP: SDP offer extracted", "from", from)

	c.mu.Lock()
	c.currentInvite = req
	c.currentTx = tx
	c.mu.Unlock()

	// Send 180 Ringing
	ringing := sipmsg.NewResponseFromRequest(req, 180, "Ringing", nil)
	if err := tx.Respond(ringing); err != nil {
		slog.ErrorContext(ctx, "SIP: failed to send 180 Ringing", "error", err)
	} else {
		slog.DebugContext(ctx, "SIP: sent 180 Ringing", "from", from)
	}

	select {
	case c.incoming <- bridge.IncomingCall{From: from, Offer: sdpJSON}:
		slog.DebugContext(ctx, "SIP: incoming call forwarded to bridge", "from", from)
	default:
		slog.WarnContext(ctx, "SIP: incoming call channel full, rejecting call", "from", from)
		busy := sipmsg.NewResponseFromRequest(req, 486, "Busy Here", nil)
		_ = tx.Respond(busy)
	}
}

func (c *Client) handleBye(ctx context.Context, req *sipmsg.Request, tx sipmsg.ServerTransaction) {
	slog.InfoContext(ctx, "SIP: 📞 BYE received (call ended)")
	ok := sipmsg.NewResponseFromRequest(req, 200, "OK", nil)
	if err := tx.Respond(ok); err != nil {
		slog.ErrorContext(ctx, "SIP: failed to send 200 OK to BYE", "error", err)
	} else {
		slog.DebugContext(ctx, "SIP: sent 200 OK to BYE")
	}
}

// AcceptCall sends a SIP 200 OK with the given SDP answer (implements bridge.SIPClient).
func (c *Client) AcceptCall(ctx context.Context, answer json.RawMessage) error {
	c.mu.Lock()
	req := c.currentInvite
	tx := c.currentTx
	c.mu.Unlock()

	if req == nil || tx == nil {
		slog.WarnContext(ctx, "SIP: attempted to accept call but no active INVITE")
		return fmt.Errorf("no active INVITE to accept")
	}

	// Unmarshal the answer SDP (stored as JSON string)
	var sdpStr string
	if err := json.Unmarshal(answer, &sdpStr); err != nil {
		// If it's not a plain JSON string, use the raw bytes as body directly
		slog.DebugContext(ctx, "SIP: SDP answer is not a JSON string, using raw bytes", "error", err)
		sdpStr = string(answer)
	}

	resp := sipmsg.NewResponseFromRequest(req, 200, "OK", []byte(sdpStr))
	resp.AppendHeader(sipmsg.NewHeader("Content-Type", "application/sdp"))
	if err := tx.Respond(resp); err != nil {
		slog.ErrorContext(ctx, "SIP: failed to send 200 OK response", "error", err)
		return fmt.Errorf("failed to send 200 OK: %w", err)
	}
	slog.InfoContext(ctx, "SIP: ✅ call accepted, 200 OK sent")
	return nil
}

// MakeCall initiates an outgoing SIP INVITE to 'to' with the given SDP offer (implements bridge.SIPClient).
func (c *Client) MakeCall(ctx context.Context, to string, offer json.RawMessage) error {
	slog.InfoContext(ctx, "SIP: initiating outgoing call", "to", to)

	if c.sipClient == nil {
		slog.WarnContext(ctx, "SIP: attempted to make call but SIP client not initialized")
		return fmt.Errorf("SIP client not initialized; call Register first")
	}

	// Extract the SDP string from the JSON-encoded offer.
	var sdpStr string
	if err := json.Unmarshal(offer, &sdpStr); err != nil {
		// Not a plain JSON string — use the raw bytes directly.
		slog.DebugContext(ctx, "SIP: SDP offer is not a JSON string, using raw bytes", "error", err)
		sdpStr = string(offer)
	}

	// Build the INVITE request.
	recipientURI := sipmsg.Uri{}
	sipmsg.ParseUri(fmt.Sprintf("sip:%s@%s", to, c.cfg.SIPServer), &recipientURI)
	req := sipmsg.NewRequest(sipmsg.INVITE, recipientURI)
	req.AppendHeader(sipmsg.NewHeader("Content-Type", "application/sdp"))
	req.SetBody([]byte(sdpStr))

	localIP := c.localIP()
	req.AppendHeader(sipmsg.NewHeader("Contact",
		fmt.Sprintf("<sip:%s@%s>", c.cfg.SIPUsername, localIP)))
	req.SetTransport("UDP")
	slog.DebugContext(ctx, "SIP: INVITE request built", "to", recipientURI.String())

	// Log outbound request details so CSeq and other headers are visible at debug level.
	hdrs := req.Headers()
	hdrStrs := make([]string, len(hdrs))
	for i, h := range hdrs {
		hdrStrs[i] = h.String()
	}
	slog.DebugContext(ctx, "SIP: outbound request headers",
		"method", req.Method,
		"uri", req.Recipient.String(),
		"headers", hdrStrs,
		"body_size", len(req.Body()),
	)

	tx, err := c.sipClient.TransactionRequest(ctx, req, sipgo.ClientRequestIncreaseCSEQ, sipgo.ClientRequestAddVia)
	if err != nil {
		slog.ErrorContext(ctx, "SIP: INVITE transaction failed", "error", err)
		return fmt.Errorf("SIP INVITE transaction failed: %w", err)
	}
	defer tx.Terminate()

	res, err := waitResponse(ctx, tx)
	if err != nil {
		slog.ErrorContext(ctx, "SIP: no response to INVITE", "error", err)
		return fmt.Errorf("SIP INVITE response error: %w", err)
	}
	slog.InfoContext(ctx, "SIP: outgoing call response received", "status_code", res.StatusCode)

	if res.StatusCode != 200 {
		slog.WarnContext(ctx, "SIP: INVITE not accepted", "status_code", res.StatusCode, "reason", res.Reason)
		return fmt.Errorf("SIP INVITE failed with status %d: %s", res.StatusCode, res.Reason)
	}

	slog.InfoContext(ctx, "SIP: ✅ outgoing call accepted", "to", to)
	return nil
}


func (c *Client) RejectCall(ctx context.Context) error {
	c.mu.Lock()
	req := c.currentInvite
	tx := c.currentTx
	c.mu.Unlock()

	if req == nil || tx == nil {
		slog.WarnContext(ctx, "SIP: attempted to reject call but no active INVITE")
		return fmt.Errorf("no active INVITE to reject")
	}
	resp := sipmsg.NewResponseFromRequest(req, 486, "Busy Here", nil)
	if err := tx.Respond(resp); err != nil {
		slog.ErrorContext(ctx, "SIP: failed to send 486 response", "error", err)
		return fmt.Errorf("failed to send 486: %w", err)
	}
	c.mu.Lock()
	c.currentInvite = nil
	c.currentTx = nil
	c.mu.Unlock()
	slog.InfoContext(ctx, "SIP: ❌ call rejected, 486 Busy sent")
	return nil
}

// HangupCall sends a SIP BYE (implements bridge.SIPClient).
func (c *Client) HangupCall(ctx context.Context) error {
	c.mu.Lock()
	c.currentInvite = nil
	c.currentTx = nil
	c.mu.Unlock()
	slog.InfoContext(ctx, "SIP: 📞 call ended (hangup)")
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
