package signaling

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 寫入等待時間
	writeWait = 10 * time.Second

	// Pong 等待時間
	pongWait = 60 * time.Second

	// Ping 間隔（必須小於 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大訊息大小
	maxMessageSize = 512 * 1024 // 512KB
)

// Client 代表一個 WebSocket 客戶端連接
type Client struct {
	// WebSocket 連接
	conn *websocket.Conn

	// 所屬的 Hub
	hub *Hub

	// 發送訊息的 channel（緩衝）
	send chan *Message

	// 客戶端所在的房間 ID
	roomID string

	// 客戶端的用戶 ID
	userID string
}

// readPump 從 WebSocket 讀取訊息並轉發到 Hub
func (c *Client) readPump() {
	defer func() {
		// 斷線時註銷客戶端
		c.hub.unregister <- c
		c.conn.Close()
		log.Printf("readPump 結束，用戶 %s 斷線", c.userID)
	}()

	// 設置讀取限制
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// 持續讀取訊息
	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 錯誤: %v", err)
			}
			break
		}

		// 確保訊息帶有房間和用戶資訊
		msg.RoomID = c.roomID
		msg.UserID = c.userID

		log.Printf("收到訊息 - 類型: %s, 房間: %s, 用戶: %s",
			msg.Type, msg.RoomID, msg.UserID)

		// 處理不同類型的訊息
		c.handleMessage(&msg)
	}
}

// writePump 從 channel 讀取訊息並發送到 WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Printf("writePump 結束，用戶 %s", c.userID)
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 關閉了 send channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 發送 JSON 訊息
			err := c.conn.WriteJSON(message)
			if err != nil {
				log.Printf("發送訊息失敗: %v", err)
				return
			}

			log.Printf("發送訊息 - 類型: %s, 目標用戶: %s",
				message.Type, c.userID)

		case <-ticker.C:
			// 定期發送 ping 保持連線
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 處理客戶端發送的訊息
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case TypeJoin:
		// 加入房間（已經在註冊時處理，這裡發送歡迎訊息）
		log.Printf("用戶 %s 發送 join 訊息", c.userID)

	case TypeOffer:
		// 處理 Offer：發送到 SFU
		c.handleOffer(msg)

	case TypeAnswer:
		// Answer 通常是 SFU 回傳的，這裡廣播（理論上不會收到客戶端的 Answer）
		c.hub.broadcast <- &BroadcastMessage{
			RoomID:  c.roomID,
			Message: msg,
			Sender:  c,
		}

	case TypeICECandidate:
		// ICE Candidate 發送到 SFU
		c.handleICECandidate(msg)

	case TypeLeave:
		// 主動離開房間
		log.Printf("用戶 %s 主動離開房間", c.userID)
		c.hub.unregister <- c

	default:
		log.Printf("未知訊息類型: %s", msg.Type)
	}
}

// handleOffer 處理 Offer，發送到 SFU 並回傳 Answer
func (c *Client) handleOffer(msg *Message) {
	log.Printf("處理 Offer - 房間: %s, 用戶: %s", c.roomID, c.userID)

	// 從 payload 提取 SDP
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		log.Printf("Offer payload 格式錯誤")
		return
	}

	sdp, ok := payload["sdp"].(string)
	if !ok {
		log.Printf("Offer SDP 格式錯誤")
		return
	}

	// 調用 SFU 處理 Offer
	answerSDP, err := c.hub.sfu.HandleOffer(c.roomID, c.userID, sdp)
	if err != nil {
		log.Printf("SFU 處理 Offer 失敗: %v", err)
		return
	}

	// 發送 Answer 回給客戶端
	answer := &Message{
		Type:   TypeAnswer,
		RoomID: c.roomID,
		UserID: "sfu", // 來自 SFU
		Payload: map[string]interface{}{
			"sdp":  answerSDP,
			"type": "answer",
		},
	}

	// 直接發送給這個客戶端
	select {
	case c.send <- answer:
		log.Printf("Answer 已發送給用戶: %s", c.userID)
	default:
		log.Printf("發送 Answer 失敗，channel 已滿")
	}
}

// handleICECandidate 處理 ICE Candidate
func (c *Client) handleICECandidate(msg *Message) {
	log.Printf("處理 ICE Candidate - 房間: %s, 用戶: %s", c.roomID, c.userID)

	// 將 ICE Candidate 轉為 JSON
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		log.Printf("序列化 ICE Candidate 失敗: %v", err)
		return
	}

	// 發送到 SFU
	if err := c.hub.sfu.HandleICECandidate(c.roomID, c.userID, string(payloadBytes)); err != nil {
		log.Printf("SFU 處理 ICE Candidate 失敗: %v", err)
	}
}
