// WebRTC Service - PeerConnection 封裝

export interface RTCConfig {
  iceServers: RTCIceServer[]
}

export class WebRTCService {
  private pc: RTCPeerConnection | null = null
  private localStream: MediaStream | null = null
  
  // 事件回調
  public onIceCandidate?: (candidate: RTCIceCandidate) => void
  public onTrack?: (event: RTCTrackEvent) => void
  public onConnectionStateChange?: (state: RTCPeerConnectionState) => void

  constructor(private config: RTCConfig) {}

  /**
   * 獲取本地媒體流（攝像頭+麥克風）
   */
  async getLocalStream(constraints?: MediaStreamConstraints): Promise<MediaStream> {
    console.log('正在獲取本地媒體流...')
    
    try {
      const defaultConstraints: MediaStreamConstraints = {
        video: {
          width: { ideal: 1280 },
          height: { ideal: 720 },
          frameRate: { ideal: 30 }
        },
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true
        }
      }

      this.localStream = await navigator.mediaDevices.getUserMedia(
        constraints || defaultConstraints
      )
      
      console.log('✅ 本地媒體流獲取成功:', this.localStream.getTracks().map(t => t.kind))
      return this.localStream
    } catch (error) {
      console.error('❌ 獲取本地媒體流失敗:', error)
      throw error
    }
  }

  /**
   * 創建 PeerConnection
   */
  createPeerConnection(): RTCPeerConnection {
    console.log('創建 PeerConnection，配置:', this.config)
    
    this.pc = new RTCPeerConnection(this.config)

    // ICE Candidate 事件
    this.pc.onicecandidate = (event) => {
      if (event.candidate) {
        console.log('🧊 收集到 ICE Candidate:', event.candidate.candidate)
        if (this.onIceCandidate) {
          this.onIceCandidate(event.candidate)
        }
      } else {
        console.log('✅ ICE Candidate 收集完成')
      }
    }

    // 接收遠端 Track 事件（觀眾端用）
    this.pc.ontrack = (event) => {
      console.log('📺 收到遠端 Track:', event.track.kind)
      if (this.onTrack) {
        this.onTrack(event)
      }
    }

    // 連接狀態變化
    this.pc.onconnectionstatechange = () => {
      const state = this.pc!.connectionState
      console.log('🔗 連接狀態變化:', state)
      if (this.onConnectionStateChange) {
        this.onConnectionStateChange(state)
      }
    }

    // ICE 連接狀態變化
    this.pc.oniceconnectionstatechange = () => {
      console.log('🧊 ICE 連接狀態:', this.pc!.iceConnectionState)
    }

    console.log('✅ PeerConnection 創建成功')
    return this.pc
  }

  /**
   * 添加本地媒體流到 PeerConnection（主播端用）
   */
  addLocalStream(stream: MediaStream): void {
    if (!this.pc) {
      throw new Error('PeerConnection 未創建')
    }

    console.log('添加本地媒體流到 PeerConnection')
    stream.getTracks().forEach((track) => {
      this.pc!.addTrack(track, stream)
      console.log(`  添加 Track: ${track.kind}`)
    })
  }

  /**
   * 創建 Offer（主播端用）
   */
  async createOffer(options?: RTCOfferOptions): Promise<RTCSessionDescriptionInit> {
    if (!this.pc) {
      throw new Error('PeerConnection 未創建')
    }

    console.log('創建 Offer...')
    
    // 預設選項：確保包含音視訊
    const defaultOptions: RTCOfferOptions = {
      offerToReceiveAudio: true,
      offerToReceiveVideo: true,
      ...options
    }
    
    const offer = await this.pc.createOffer(defaultOptions)
    await this.pc.setLocalDescription(offer)
    console.log('✅ Offer 創建成功，SDP 長度:', offer.sdp?.length)
    
    return offer
  }

  /**
   * 創建 Answer（觀眾端用）
   */
  async createAnswer(): Promise<RTCSessionDescriptionInit> {
    if (!this.pc) {
      throw new Error('PeerConnection 未創建')
    }

    console.log('創建 Answer...')
    const answer = await this.pc.createAnswer()
    await this.pc.setLocalDescription(answer)
    console.log('✅ Answer 創建成功，SDP 長度:', answer.sdp?.length)
    
    return answer
  }

  /**
   * 處理遠端 SDP（Offer 或 Answer）
   */
  async setRemoteDescription(sdp: RTCSessionDescriptionInit): Promise<void> {
    if (!this.pc) {
      throw new Error('PeerConnection 未創建')
    }

    console.log(`設置遠端 SDP (${sdp.type})...`)
    await this.pc.setRemoteDescription(new RTCSessionDescription(sdp))
    console.log('✅ 遠端 SDP 設置成功')
  }

  /**
   * 添加 ICE Candidate
   */
  async addIceCandidate(candidate: RTCIceCandidateInit): Promise<void> {
    if (!this.pc) {
      throw new Error('PeerConnection 未創建')
    }

    try {
      await this.pc.addIceCandidate(new RTCIceCandidate(candidate))
      console.log('✅ ICE Candidate 添加成功')
    } catch (error) {
      console.error('❌ 添加 ICE Candidate 失敗:', error)
    }
  }

  /**
   * 關閉連接並清理資源
   */
  close(): void {
    console.log('關閉 WebRTC 連接...')

    // 停止本地媒體流
    if (this.localStream) {
      this.localStream.getTracks().forEach((track) => {
        track.stop()
        console.log(`  停止 Track: ${track.kind}`)
      })
      this.localStream = null
    }

    // 關閉 PeerConnection
    if (this.pc) {
      this.pc.close()
      this.pc = null
    }

    console.log('✅ WebRTC 連接已關閉')
  }

  /**
   * 獲取連接統計資訊（調試用）
   */
  async getStats(): Promise<RTCStatsReport | null> {
    if (!this.pc) {
      return null
    }
    return await this.pc.getStats()
  }

  /**
   * 獲取 PeerConnection 實例
   */
  getPeerConnection(): RTCPeerConnection | null {
    return this.pc
  }

  /**
   * 獲取本地媒體流
   */
  getLocalMediaStream(): MediaStream | null {
    return this.localStream
  }
}

