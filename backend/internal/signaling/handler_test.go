package signaling

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockRouter records all routed messages.
type mockRouter struct {
	messages []Message
}

func (m *mockRouter) Route(msg Message, send func(Message) error) {
	m.messages = append(m.messages, msg)
}

func TestHub_Broadcast(t *testing.T) {
	router := &mockRouter{}
	hub := NewHub(router)

	// Start a test WebSocket server using the hub.
	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Read the initial status message
	var statusMsg Message
	if err := conn.ReadJSON(&statusMsg); err != nil {
		t.Fatalf("read initial status: %v", err)
	}
	if statusMsg.Type != TypeStatus {
		t.Errorf("expected status type, got %q", statusMsg.Type)
	}

	// Broadcast an incoming message
	hub.Broadcast(Message{Type: TypeIncoming, Payload: map[string]interface{}{"from": "alice"}})

	// Expect the broadcast to arrive at the client
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var rcv Message
	if err := conn.ReadJSON(&rcv); err != nil {
		t.Fatalf("read broadcast: %v", err)
	}
	if rcv.Type != TypeIncoming {
		t.Errorf("expected incoming type, got %q", rcv.Type)
	}
}

func TestHub_RouteToRouter(t *testing.T) {
	router := &mockRouter{}
	hub := NewHub(router)

	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Consume the initial status message
	var firstMsg Message
	conn.ReadJSON(&firstMsg)

	// Send a message from client → hub → router
	sent := Message{Type: TypeAnswer, Data: "test-answer"}
	if err := conn.WriteJSON(sent); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // allow the read loop to process

	if len(router.messages) == 0 {
		t.Fatal("expected router to receive a message")
	}
	if router.messages[0].Type != TypeAnswer {
		t.Errorf("expected answer type, got %q", router.messages[0].Type)
	}
}

func TestHub_SetRouter(t *testing.T) {
	hub := NewHub(nil)
	router := &mockRouter{}
	hub.SetRouter(router)

	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Consume initial status
	var firstMsg2 Message
	conn.ReadJSON(&firstMsg2)

	if err := conn.WriteJSON(Message{Type: TypeHangup}); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if len(router.messages) == 0 || router.messages[0].Type != TypeHangup {
		t.Errorf("expected hangup to be routed, got %v", router.messages)
	}
}
