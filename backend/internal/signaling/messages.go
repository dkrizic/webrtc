package signaling

// Message types for WebSocket signaling
type MessageType string

const (
	TypeOffer    MessageType = "offer"
	TypeAnswer   MessageType = "answer"
	TypeICE      MessageType = "ice"
	TypeDial     MessageType = "dial"
	TypeHangup   MessageType = "hangup"
	TypeStatus   MessageType = "status"
	TypeError    MessageType = "error"
	TypeIncoming MessageType = "incoming"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Data    string      `json:"data,omitempty"`
}
