// Signaling Service - WebSocket 通訊封裝

export interface Message {
  type: 'offer' | 'answer' | 'ice_candidate' | 'join' | 'leave'
  room_id: string
  user_id: string
  payload: any
}

export class SignalingService {
  private ws: WebSocket | null = null
  private messageHandlers: Map<string, (data: any) => void> = new Map()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 2000 // 2 秒

  /**
   * 連接到 Signaling Server
   * @param url WebSocket URL (例如: ws://localhost:8080/ws?room_id=xxx&user_id=xxx)
   */
  connect(url: string): Promise<void> {
    return new Promise((resolve, reject) => {
      console.log('正在連接 Signaling Server:', url)

      try {
        this.ws = new WebSocket(url)

        this.ws.onopen = () => {
          console.log('✅ WebSocket 連接成功')
          this.reconnectAttempts = 0
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message: Message = JSON.parse(event.data)
            console.log('📨 收到訊息:', message.type, message)
            
            // 調用對應的處理器
            const handler = this.messageHandlers.get(message.type)
            if (handler) {
              handler(message)
            } else {
              console.warn('沒有處理器處理訊息類型:', message.type)
            }
          } catch (error) {
            console.error('解析訊息失敗:', error)
          }
        }

        this.ws.onerror = (error) => {
          console.error('❌ WebSocket 錯誤:', error)
          reject(error)
        }

        this.ws.onclose = (event) => {
          console.log('WebSocket 連接關閉:', event.code, event.reason)
          
          // 嘗試重連
          if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++
            console.log(`嘗試重連 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`)
            setTimeout(() => {
              this.connect(url)
            }, this.reconnectDelay)
          }
        }
      } catch (error) {
        console.error('創建 WebSocket 失敗:', error)
        reject(error)
      }
    })
  }

  /**
   * 發送訊息到 Signaling Server
   */
  send(message: Message): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('WebSocket 未連接，無法發送訊息')
      return
    }

    try {
      const jsonStr = JSON.stringify(message)
      this.ws.send(jsonStr)
      console.log('📤 發送訊息:', message.type, message)
    } catch (error) {
      console.error('發送訊息失敗:', error)
    }
  }

  /**
   * 註冊訊息處理器
   * @param type 訊息類型
   * @param handler 處理函數
   */
  on(type: string, handler: (data: any) => void): void {
    this.messageHandlers.set(type, handler)
    console.log(`註冊處理器: ${type}`)
  }

  /**
   * 移除訊息處理器
   */
  off(type: string): void {
    this.messageHandlers.delete(type)
  }

  /**
   * 斷開 WebSocket 連接
   */
  disconnect(): void {
    if (this.ws) {
      console.log('主動斷開 WebSocket 連接')
      this.reconnectAttempts = this.maxReconnectAttempts // 停止重連
      this.ws.close()
      this.ws = null
    }
  }

  /**
   * 檢查連接狀態
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

