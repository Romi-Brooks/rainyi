package service

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	UserID int64
}

type WebSocketHub struct {
	clients map[int64]*Client
	mu      chan struct{}
}

type wsMessage struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients: make(map[int64]*Client),
		mu:      make(chan struct{}, 1),
	}
}

func (h *WebSocketHub) lock() {
	h.mu <- struct{}{}
}

func (h *WebSocketHub) unlock() {
	<-h.mu
}

// TODO: 分布式部署时 WebSocketHub 需迁移到 Redis Pub/Sub
// 1. 发消息时 publish 到 Redis channel "ws:notify:{userID}"
// 2. 每台机器订阅自己的 channel
// 3. 收到消息后查询本地 clients map 推送
// 4. 进程重启后需从 Redis 恢复心跳
func (h *WebSocketHub) Register(client *Client) {
	h.lock()
	defer h.unlock()
	if existing, ok := h.clients[client.UserID]; ok {
		close(existing.Send)
		existing.Conn.Close()
	}
	h.clients[client.UserID] = client
}

func (h *WebSocketHub) Unregister(client *Client) {
	h.lock()
	defer h.unlock()
	if existing, ok := h.clients[client.UserID]; ok && existing == client {
		delete(h.clients, client.UserID)
		close(client.Send)
	}
}

func (h *WebSocketHub) GetClient(userID int64) *Client {
	h.lock()
	defer h.unlock()
	return h.clients[userID]
}

func (h *WebSocketHub) SendToUser(userID int64, message interface{}) {
	client := h.GetClient(userID)
	if client == nil {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("WebSocket marshal message error: %v", err)
		return
	}

	select {
	case client.Send <- data:
	default:
		log.Printf("WebSocket send channel full for user %d, dropping message", userID)
	}
}

func (h *WebSocketHub) BroadcastStarting(userID int64) {
	h.SendToUser(userID, wsMessage{Type: "ai_start"})
}

func (h *WebSocketHub) SendStreamChunk(userID int64, content string) {
	h.SendToUser(userID, wsMessage{Type: "stream", Content: content})
}

func (h *WebSocketHub) SendComplete(userID int64, fullContent string) {
	h.SendToUser(userID, wsMessage{Type: "complete", Content: fullContent})
}

func (h *WebSocketHub) SendError(userID int64, errMsg string) {
	h.SendToUser(userID, wsMessage{Type: "error", Content: errMsg})
}
