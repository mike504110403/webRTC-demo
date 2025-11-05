package signaling

import (
	"log"
	"sync"

	"github.com/mike504110403/webrtc-demo/internal/sfu"
)

// Hub 管理所有 WebSocket 連接和房間
type Hub struct {
	// 房間映射：roomID -> 房間內的所有客戶端
	rooms map[string]map[*Client]bool

	// 註冊新客戶端
	register chan *Client

	// 註銷客戶端
	unregister chan *Client

	// 廣播訊息到特定房間
	broadcast chan *BroadcastMessage

	// 互斥鎖保護 rooms map
	mu sync.RWMutex

	// SFU 實例
	sfu *sfu.SFU
}

// BroadcastMessage 廣播訊息結構
type BroadcastMessage struct {
	RoomID  string
	Message *Message
	Sender  *Client // 發送者（可選，用於排除發送者自己）
}

// NewHub 創建新的 Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256), // 緩衝區避免阻塞
		sfu:        sfu.NewSFU(),                      // 初始化 SFU
	}
}

// Run 啟動 Hub 主循環（處理註冊/註銷/廣播）
func (h *Hub) Run() {
	log.Println("Hub 啟動，開始處理連接...")

	for {
		select {
		case client := <-h.register:
			// 註冊新客戶端到房間
			h.registerClient(client)

		case client := <-h.unregister:
			// 從房間註銷客戶端
			h.unregisterClient(client)

		case broadcastMsg := <-h.broadcast:
			// 廣播訊息到房間內所有客戶端
			h.broadcastToRoom(broadcastMsg)
		}
	}
}

// registerClient 註冊客戶端到房間
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果房間不存在，創建新房間
	if h.rooms[client.roomID] == nil {
		h.rooms[client.roomID] = make(map[*Client]bool)
		log.Printf("創建新房間: %s", client.roomID)
	}

	// 將客戶端加入房間
	h.rooms[client.roomID][client] = true

	log.Printf("用戶 %s 加入房間 %s，當前房間人數: %d",
		client.userID, client.roomID, len(h.rooms[client.roomID]))
}

// unregisterClient 從房間註銷客戶端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.roomID]; ok {
		if _, exists := clients[client]; exists {
			// 從房間移除客戶端
			delete(clients, client)

			// 關閉發送通道
			close(client.send)

			// 通知 SFU 移除 Peer
			h.sfu.RemovePeer(client.roomID, client.userID)

			log.Printf("用戶 %s 離開房間 %s，剩餘人數: %d",
				client.userID, client.roomID, len(clients))

			// 如果房間空了，刪除房間
			if len(clients) == 0 {
				delete(h.rooms, client.roomID)
				log.Printf("房間 %s 已清空，刪除房間", client.roomID)
			}
		}
	}
}

// broadcastToRoom 廣播訊息到房間內所有客戶端
func (h *Hub) broadcastToRoom(broadcastMsg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[broadcastMsg.RoomID]
	if !ok {
		log.Printf("房間 %s 不存在，無法廣播", broadcastMsg.RoomID)
		return
	}

	// 發送給房間內所有客戶端（除了發送者自己）
	sentCount := 0
	for client := range clients {
		// 如果有指定發送者，則不發送給發送者自己
		if broadcastMsg.Sender != nil && client == broadcastMsg.Sender {
			continue
		}

		select {
		case client.send <- broadcastMsg.Message:
			sentCount++
		default:
			// 發送失敗，客戶端可能已斷線
			log.Printf("發送訊息給 %s 失敗，客戶端可能已斷線", client.userID)
		}
	}

	log.Printf("廣播訊息到房間 %s，類型: %s，成功發送: %d/%d",
		broadcastMsg.RoomID, broadcastMsg.Message.Type, sentCount, len(clients))
}

// GetRoomClients 獲取房間內所有客戶端（用於調試）
func (h *Hub) GetRoomClients(roomID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[roomID]
	if !ok {
		return nil
	}

	result := make([]*Client, 0, len(clients))
	for client := range clients {
		result = append(result, client)
	}

	return result
}
