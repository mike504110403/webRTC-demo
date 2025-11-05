package sfu

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/pion/webrtc/v3"
)

// SFU 簡單的 SFU 實現（1對多轉發）
type SFU struct {
	// 房間映射：roomID -> Room
	rooms map[string]*Room
	mu    sync.RWMutex

	// WebRTC 配置
	config webrtc.Configuration
}

// Room 房間，包含一個 Publisher 和多個 Subscribers
type Room struct {
	ID          string
	Publisher   *Peer            // 主播
	Subscribers map[string]*Peer // 觀眾
	mu          sync.RWMutex
}

// Peer 代表一個 WebRTC 連接
type Peer struct {
	ID             string
	RoomID         string
	PeerConnection *webrtc.PeerConnection
	Tracks         []*webrtc.TrackLocalStaticRTP // 用於轉發的本地 Track
}

// NewSFU 創建新的 SFU
func NewSFU() *SFU {
	return &SFU{
		rooms: make(map[string]*Room),
		config: webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		},
	}
}

// HandleOffer 處理 Offer（主播或觀眾發來的）
func (s *SFU) HandleOffer(roomID, userID string, offerSDP string) (string, error) {
	s.mu.Lock()

	// 確保房間存在
	room, exists := s.rooms[roomID]
	if !exists {
		room = &Room{
			ID:          roomID,
			Subscribers: make(map[string]*Peer),
		}
		s.rooms[roomID] = room
		log.Printf("[SFU] 創建房間: %s", roomID)
	}
	s.mu.Unlock()

	// 判斷是主播還是觀眾
	// 簡單判斷：如果房間沒有 Publisher，就是主播
	isPublisher := room.Publisher == nil

	if isPublisher {
		return s.handlePublisherOffer(room, userID, offerSDP)
	} else {
		return s.handleSubscriberOffer(room, userID, offerSDP)
	}
}

// handlePublisherOffer 處理主播的 Offer
func (s *SFU) handlePublisherOffer(room *Room, userID string, offerSDP string) (string, error) {
	log.Printf("[SFU] 處理主播 Offer，房間: %s, 用戶: %s", room.ID, userID)

	// 創建 PeerConnection
	pc, err := webrtc.NewPeerConnection(s.config)
	if err != nil {
		return "", err
	}

	// 創建 Peer
	peer := &Peer{
		ID:             userID,
		RoomID:         room.ID,
		PeerConnection: pc,
		Tracks:         make([]*webrtc.TrackLocalStaticRTP, 0),
	}

	// 監聽 Track（主播推送的音視訊）
	pc.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("[SFU] 收到主播 Track: %s, 類型: %s", remoteTrack.ID(), remoteTrack.Kind())

		// 創建本地 Track 用於轉發給觀眾
		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			remoteTrack.ID(),
			remoteTrack.StreamID(),
		)
		if err != nil {
			log.Printf("創建本地 Track 失敗: %v", err)
			return
		}

		peer.Tracks = append(peer.Tracks, localTrack)

		// 從遠端 Track 讀取 RTP 封包並轉發到本地 Track
		go func() {
			rtpBuf := make([]byte, 1400)
			for {
				n, _, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					log.Printf("讀取 RTP 封包失敗: %v", readErr)
					return
				}

				// 寫入本地 Track（會自動轉發給所有訂閱者）
				if _, writeErr := localTrack.Write(rtpBuf[:n]); writeErr != nil {
					log.Printf("寫入 RTP 封包失敗: %v", writeErr)
					return
				}
			}
		}()

		// 將這個 Track 添加到所有已存在的觀眾連接中
		room.mu.RLock()
		for _, subscriber := range room.Subscribers {
			if _, err := subscriber.PeerConnection.AddTrack(localTrack); err != nil {
				log.Printf("添加 Track 到訂閱者失敗: %v", err)
			} else {
				log.Printf("[SFU] 添加 Track 到訂閱者: %s", subscriber.ID)
			}
		}
		room.mu.RUnlock()
	})

	// ICE 連接狀態監控
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("[SFU] 主播 ICE 狀態: %s", state.String())
	})

	// 設置遠端 SDP
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}
	if err := pc.SetRemoteDescription(offer); err != nil {
		return "", err
	}

	// 創建 Answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	// 設置本地 SDP
	if err := pc.SetLocalDescription(answer); err != nil {
		return "", err
	}

	// 保存 Publisher
	room.mu.Lock()
	room.Publisher = peer
	room.mu.Unlock()

	log.Printf("[SFU] 主播連接建立成功，房間: %s", room.ID)
	return answer.SDP, nil
}

// handleSubscriberOffer 處理觀眾的 Offer
func (s *SFU) handleSubscriberOffer(room *Room, userID string, offerSDP string) (string, error) {
	log.Printf("[SFU] 處理觀眾 Offer，房間: %s, 用戶: %s", room.ID, userID)

	// 創建 PeerConnection
	pc, err := webrtc.NewPeerConnection(s.config)
	if err != nil {
		return "", err
	}

	// 添加主播的所有 Tracks 到這個觀眾的 PeerConnection
	room.mu.RLock()
	if room.Publisher != nil {
		for _, track := range room.Publisher.Tracks {
			if _, err := pc.AddTrack(track); err != nil {
				log.Printf("添加 Track 到觀眾失敗: %v", err)
			} else {
				log.Printf("[SFU] 添加主播 Track 到觀眾: %s", userID)
			}
		}
	}
	room.mu.RUnlock()

	// ICE 連接狀態監控
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("[SFU] 觀眾 %s ICE 狀態: %s", userID, state.String())
	})

	// 設置遠端 SDP
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}
	if err := pc.SetRemoteDescription(offer); err != nil {
		return "", err
	}

	// 創建 Answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	// 設置本地 SDP
	if err := pc.SetLocalDescription(answer); err != nil {
		return "", err
	}

	// 創建並保存 Peer
	peer := &Peer{
		ID:             userID,
		RoomID:         room.ID,
		PeerConnection: pc,
	}

	room.mu.Lock()
	room.Subscribers[userID] = peer
	room.mu.Unlock()

	log.Printf("[SFU] 觀眾連接建立成功，房間: %s, 觀眾: %s", room.ID, userID)
	return answer.SDP, nil
}

// HandleICECandidate 處理 ICE Candidate
func (s *SFU) HandleICECandidate(roomID, userID string, candidateJSON string) error {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		log.Printf("[SFU] 房間不存在: %s", roomID)
		return nil
	}

	// 找到對應的 Peer
	var peer *Peer
	room.mu.RLock()
	if room.Publisher != nil && room.Publisher.ID == userID {
		peer = room.Publisher
	} else {
		peer = room.Subscribers[userID]
	}
	room.mu.RUnlock()

	if peer == nil {
		log.Printf("[SFU] Peer 不存在: %s", userID)
		return nil
	}

	// 解析 ICE Candidate
	var candidate webrtc.ICECandidateInit
	if err := json.Unmarshal([]byte(candidateJSON), &candidate); err != nil {
		return err
	}

	// 添加 ICE Candidate
	if err := peer.PeerConnection.AddICECandidate(candidate); err != nil {
		return err
	}

	log.Printf("[SFU] 添加 ICE Candidate 成功: %s", userID)
	return nil
}

// RemovePeer 移除 Peer
func (s *SFU) RemovePeer(roomID, userID string) {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// 如果是主播離開
	if room.Publisher != nil && room.Publisher.ID == userID {
		room.Publisher.PeerConnection.Close()
		room.Publisher = nil
		log.Printf("[SFU] 主播離開房間: %s", roomID)

		// 關閉所有觀眾連接
		for _, subscriber := range room.Subscribers {
			subscriber.PeerConnection.Close()
		}
		room.Subscribers = make(map[string]*Peer)

		// 刪除房間
		s.mu.Lock()
		delete(s.rooms, roomID)
		s.mu.Unlock()
		log.Printf("[SFU] 房間已刪除: %s", roomID)
		return
	}

	// 如果是觀眾離開
	if subscriber, exists := room.Subscribers[userID]; exists {
		subscriber.PeerConnection.Close()
		delete(room.Subscribers, userID)
		log.Printf("[SFU] 觀眾離開房間: %s, 用戶: %s", roomID, userID)
	}
}
