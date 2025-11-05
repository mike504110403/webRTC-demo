package signaling

// MessageType 訊息類型
type MessageType string

const (
	// WebRTC Signaling
	TypeOffer        MessageType = "offer"
	TypeAnswer       MessageType = "answer"
	TypeICECandidate MessageType = "ice_candidate"
	
	// 連線管理
	TypeJoin  MessageType = "join"
	TypeLeave MessageType = "leave"
)

// Message WebSocket 訊息格式
type Message struct {
	Type    MessageType `json:"type"`
	RoomID  string      `json:"room_id"`
	UserID  string      `json:"user_id"`
	Payload interface{} `json:"payload"`
}

// SDPPayload SDP 訊息
type SDPPayload struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"` // "offer" or "answer"
}

// ICEPayload ICE Candidate 訊息
type ICEPayload struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}

