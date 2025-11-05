package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mike504110403/webrtc-demo/internal/signaling"
	"github.com/rs/cors"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("========================================")
	log.Printf("  WebRTC 直播系統 - Signaling Server")
	log.Printf("========================================")
	log.Printf("監聽端口: %s", port)
	log.Printf("架構: P2SFU2P（內建 Pion WebRTC SFU）")
	log.Printf("")

	// 初始化 Signaling Hub（包含內建 SFU）
	hub := signaling.NewHub()
	signaling.GlobalHub = hub
	go hub.Run()
	log.Printf("✓ Signaling Hub 已啟動")
	log.Printf("✓ 內建 SFU 已就緒")
	log.Printf("")

	// 設置路由
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", signaling.HandleWebSocket)

	// 健康檢查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// CORS 設置（允許局域網訪問）
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // 允許所有來源
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false, // 配合 * 使用
	}).Handler(mux)

	log.Printf("========================================")
	log.Printf("服務已啟動:")
	log.Printf("  • 本機:      http://localhost:%s", port)
	log.Printf("  • 局域網:    http://192.168.1.181:%s", port)
	log.Printf("  • WebSocket: ws://[host]:%s/ws", port)
	log.Printf("========================================")
	log.Printf("")

	// 監聽所有網路介面（允許局域網訪問）
	if err := http.ListenAndServe("0.0.0.0:"+port, handler); err != nil {
		log.Fatal("服務啟動失敗:", err)
	}
}

