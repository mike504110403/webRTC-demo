package signaling

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	// WebSocket 升級器
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// 允許所有來源（開發階段，生產環境需要限制）
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// 全局 Hub 實例（由 main.go 初始化）
	GlobalHub *Hub
)

// HandleWebSocket 處理 WebSocket 連接
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("收到 WebSocket 連接請求，來源: %s", r.RemoteAddr)

	// 從 URL 參數獲取房間 ID 和用戶 ID
	roomID := r.URL.Query().Get("room_id")
	userID := r.URL.Query().Get("user_id")

	// 驗證參數
	if roomID == "" || userID == "" {
		log.Println("缺少必要參數: room_id 或 user_id")
		http.Error(w, "缺少參數 room_id 或 user_id", http.StatusBadRequest)
		return
	}

	// 升級 HTTP 連接到 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升級失敗: %v", err)
		return
	}

	// 創建客戶端
	client := &Client{
		conn:   conn,
		hub:    GlobalHub,
		send:   make(chan *Message, 256), // 緩衝 channel
		roomID: roomID,
		userID: userID,
	}

	// 註冊客戶端到 Hub
	client.hub.register <- client

	log.Printf("WebSocket 連接建立成功 - 房間: %s, 用戶: %s", roomID, userID)

	// 啟動讀寫協程
	go client.writePump() // 先啟動寫入協程
	go client.readPump()  // 再啟動讀取協程（會阻塞直到斷線）
}
