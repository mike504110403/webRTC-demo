<template>
  <div class="broadcaster">
    <h1>主播端（推流）</h1>
    
    <div class="video-container">
      <video ref="localVideo" autoplay muted playsinline></video>
    </div>

    <div class="controls">
      <button @click="startBroadcast">開始直播</button>
      <button @click="stopBroadcast">停止直播</button>
    </div>

    <div class="status">
      <p>狀態: {{ status }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { SignalingService } from '../services/signaling'
import { WebRTCService } from '../services/webrtc'

const localVideo = ref<HTMLVideoElement>()
const status = ref('未開始')

// 配置（支援局域網訪問）
const ROOM_ID = 'room1'
const USER_ID = 'broadcaster_' + Date.now()

// 自動檢測主機，支援局域網訪問
const host = window.location.hostname // localhost 或 192.168.1.181
const SIGNALING_URL = `ws://${host}:8080/ws?room_id=${ROOM_ID}&user_id=${USER_ID}`

console.log('Signaling URL:', SIGNALING_URL)

// 服務實例
let signalingService: SignalingService | null = null
let webrtcService: WebRTCService | null = null

/**
 * 開始直播
 */
const startBroadcast = async () => {
  try {
    status.value = '正在初始化...'
    
    // 1. 創建 WebRTC 服務
    webrtcService = new WebRTCService({
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' }
      ]
    })
    
    // 2. 獲取本地媒體流
    status.value = '正在獲取攝像頭權限...'
    const stream = await webrtcService.getLocalStream()
    
    // 3. 顯示在 video 元素
    if (localVideo.value) {
      localVideo.value.srcObject = stream
    }
    
    // 4. 連接 Signaling Server
    status.value = '正在連接 Signaling Server...'
    signalingService = new SignalingService()
    await signalingService.connect(SIGNALING_URL)
    
    // 5. 創建 PeerConnection
    status.value = '正在建立 WebRTC 連接...'
    webrtcService.createPeerConnection()
    
    // 6. 添加本地媒體流到 PeerConnection
    webrtcService.addLocalStream(stream)
    
    // 7. 設置 ICE Candidate 處理
    webrtcService.onIceCandidate = (candidate) => {
      console.log('發送 ICE Candidate')
      signalingService!.send({
        type: 'ice_candidate',
        room_id: ROOM_ID,
        user_id: USER_ID,
        payload: {
          candidate: candidate.candidate,
          sdpMid: candidate.sdpMid,
          sdpMLineIndex: candidate.sdpMLineIndex
        }
      })
    }
    
    // 8. 監聽 Answer（從 SFU 來的）
    signalingService.on('answer', async (message) => {
      console.log('收到 Answer')
      await webrtcService!.setRemoteDescription(message.payload)
    })
    
    // 9. 監聽遠端 ICE Candidate
    signalingService.on('ice_candidate', async (message) => {
      console.log('收到遠端 ICE Candidate')
      await webrtcService!.addIceCandidate(message.payload)
    })
    
    // 10. 監聽連接狀態
    webrtcService.onConnectionStateChange = (state) => {
      console.log('連接狀態:', state)
      if (state === 'connected') {
        status.value = '✅ 直播中'
      } else if (state === 'failed' || state === 'disconnected') {
        status.value = '❌ 連接失敗'
      }
    }
    
    // 11. 創建並發送 Offer
    status.value = '正在發送 Offer...'
    const offer = await webrtcService.createOffer()
    signalingService.send({
      type: 'offer',
      room_id: ROOM_ID,
      user_id: USER_ID,
      payload: {
        sdp: offer.sdp,
        type: offer.type
      }
    })
    
    status.value = '等待連接...'
    console.log('✅ 直播初始化完成，等待觀眾加入')
    
  } catch (error) {
    console.error('開始直播失敗:', error)
    status.value = '❌ 啟動失敗: ' + (error as Error).message
    stopBroadcast()
  }
}

/**
 * 停止直播
 */
const stopBroadcast = () => {
  console.log('停止直播')
  
  // 關閉 WebRTC
  if (webrtcService) {
    webrtcService.close()
    webrtcService = null
  }
  
  // 關閉 WebSocket
  if (signalingService) {
    signalingService.disconnect()
    signalingService = null
  }
  
  // 清除 video 元素
  if (localVideo.value) {
    localVideo.value.srcObject = null
  }
  
  status.value = '已停止'
}

// 組件卸載時清理
onUnmounted(() => {
  stopBroadcast()
})
</script>

<style scoped>
.broadcaster {
  padding: 20px;
}

.video-container {
  margin: 20px 0;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

video {
  width: 100%;
  max-width: 800px;
  height: auto;
}

.controls {
  display: flex;
  gap: 10px;
  margin: 20px 0;
}

button {
  padding: 10px 20px;
  font-size: 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  background: #007bff;
  color: white;
}

button:hover {
  background: #0056b3;
}

.status {
  margin-top: 20px;
  padding: 10px;
  background: #f0f0f0;
  border-radius: 4px;
}
</style>

