<template>
  <div class="viewer">
    <h1>觀眾端（拉流）</h1>
    
    <div class="video-container">
      <video ref="remoteVideo" autoplay playsinline></video>
    </div>

    <div class="controls">
      <button @click="joinStream">加入直播</button>
      <button @click="leaveStream">離開直播</button>
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

const remoteVideo = ref<HTMLVideoElement>()
const status = ref('未連接')

// 配置（支援局域網訪問）
const ROOM_ID = 'room1'
const USER_ID = 'viewer_' + Date.now()

// 自動檢測主機，支援局域網訪問
const host = window.location.hostname // localhost 或 192.168.1.181
const SIGNALING_URL = `ws://${host}:8080/ws?room_id=${ROOM_ID}&user_id=${USER_ID}`

console.log('Signaling URL:', SIGNALING_URL)

// 服務實例
let signalingService: SignalingService | null = null
let webrtcService: WebRTCService | null = null

/**
 * 加入直播
 */
const joinStream = async () => {
  try {
    status.value = '正在初始化...'
    
    // 1. 創建 WebRTC 服務
    webrtcService = new WebRTCService({
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' }
      ]
    })
    
    // 2. 連接 Signaling Server
    status.value = '正在連接 Signaling Server...'
    signalingService = new SignalingService()
    await signalingService.connect(SIGNALING_URL)
    
    // 3. 創建 PeerConnection
    status.value = '正在建立 WebRTC 連接...'
    const pc = webrtcService.createPeerConnection()
    
    // 3.5 添加 transceiver 表明想接收音視訊（重要！）
    pc.addTransceiver('video', { direction: 'recvonly' })
    pc.addTransceiver('audio', { direction: 'recvonly' })
    console.log('已添加 transceiver: video, audio (recvonly)')
    
    // 4. 監聽遠端 Track（接收主播的視訊流）
    webrtcService.onTrack = (event) => {
      console.log('📺 收到遠端媒體流')
      if (remoteVideo.value && event.streams[0]) {
        remoteVideo.value.srcObject = event.streams[0]
        status.value = '✅ 正在觀看直播'
      }
    }
    
    // 5. 設置 ICE Candidate 處理
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
    
    // 6. 監聽 Offer（如果主播先發 Offer）
    signalingService.on('offer', async (message) => {
      console.log('收到 Offer')
      await webrtcService!.setRemoteDescription(message.payload)
      
      // 創建並發送 Answer
      const answer = await webrtcService!.createAnswer()
      signalingService!.send({
        type: 'answer',
        room_id: ROOM_ID,
        user_id: USER_ID,
        payload: {
          sdp: answer.sdp,
          type: answer.type
        }
      })
    })
    
    // 7. 監聽 Answer（如果觀眾先發 Offer）
    signalingService.on('answer', async (message) => {
      console.log('收到 Answer')
      await webrtcService!.setRemoteDescription(message.payload)
    })
    
    // 8. 監聽遠端 ICE Candidate
    signalingService.on('ice_candidate', async (message) => {
      console.log('收到遠端 ICE Candidate')
      await webrtcService!.addIceCandidate(message.payload)
    })
    
    // 9. 監聽連接狀態
    webrtcService.onConnectionStateChange = (state) => {
      console.log('連接狀態:', state)
      if (state === 'connected') {
        status.value = '✅ 正在觀看直播'
      } else if (state === 'failed' || state === 'disconnected') {
        status.value = '❌ 連接失敗'
      }
    }
    
    // 10. 創建並發送 Offer（觀眾主動訂閱）
    status.value = '正在請求視訊流...'
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
    console.log('✅ 觀眾端初始化完成，等待接收視訊流')
    
  } catch (error) {
    console.error('加入直播失敗:', error)
    status.value = '❌ 加入失敗: ' + (error as Error).message
    leaveStream()
  }
}

/**
 * 離開直播
 */
const leaveStream = () => {
  console.log('離開直播')
  
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
  if (remoteVideo.value) {
    remoteVideo.value.srcObject = null
  }
  
  status.value = '未連接'
}

// 組件卸載時清理
onUnmounted(() => {
  leaveStream()
})
</script>

<style scoped>
.viewer {
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
  background: #28a745;
  color: white;
}

button:hover {
  background: #1e7e34;
}

.status {
  margin-top: 20px;
  padding: 10px;
  background: #f0f0f0;
  border-radius: 4px;
}
</style>

