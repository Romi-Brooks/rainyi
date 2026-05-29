package controller

import (
	"log"
	"net/http"
	"strings"
	"time"

	"rain-yi-backend/config"
	"rain-yi-backend/middleware"
	"rain-yi-backend/model"
	"rain-yi-backend/repository"
	"rain-yi-backend/service"
	"rain-yi-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type ChatController struct {
	msgRepo        *repository.MessageRepository
	convRepo       *repository.ConversationRepository
	aiService      *service.AIService
	contextManager *service.ContextManager
	hub            *service.WebSocketHub
	upgrader       websocket.Upgrader
}

func NewChatController(
	msgRepo *repository.MessageRepository,
	convRepo *repository.ConversationRepository,
	aiService *service.AIService,
	contextManager *service.ContextManager,
	hub *service.WebSocketHub,
) *ChatController {
	return &ChatController{
		msgRepo:        msgRepo,
		convRepo:       convRepo,
		aiService:      aiService,
		contextManager: contextManager,
		hub:            hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

type wsIncomingMessage struct {
	ConversationID int64  `json:"conversation_id"`
	Content        string `json:"content"`
}

func (ctl *ChatController) HandleWebSocket(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if tokenStr == "" || tokenStr == c.GetHeader("Authorization") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
		return
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效"})
		return
	}

	conn, err := ctl.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &service.Client{
		ID:     claims.Email,
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	ctl.hub.Register(client)

	go ctl.readPump(client)
	go ctl.writePump(client)
}

func (ctl *ChatController) readPump(client *service.Client) {
	defer func() {
		ctl.hub.Unregister(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(4096)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg wsIncomingMessage
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		content := utils.SanitizeContent(msg.Content)
		userID := client.UserID

		var conv *model.Conversation
		var findErr error
		if msg.ConversationID > 0 {
			conv, findErr = ctl.convRepo.FindByID(msg.ConversationID)
			if findErr != nil || conv.UserID != userID {
				log.Printf("Conversation not found or not owned: %d", msg.ConversationID)
				ctl.hub.SendError(userID, "会话不存在")
				continue
			}
		} else {
			conv, findErr = ctl.convRepo.FindOrCreateDefault(userID)
			if findErr != nil {
				log.Printf("FindOrCreateDefault error: %v", findErr)
				ctl.hub.SendError(userID, "会话创建失败")
				continue
			}
		}

		userMessage := &model.Message{
			ConversationID: conv.ID,
			Role:           "user",
			Content:        content,
		}
		if err := ctl.msgRepo.Create(userMessage); err != nil {
			log.Printf("Save user message error: %v", err)
			ctl.hub.SendError(userID, "消息保存失败")
			continue
		}

		history, err := ctl.contextManager.BuildContext(conv.ID)
		if err != nil {
			log.Printf("BuildContext error: %v", err)
			ctl.hub.SendError(userID, "上下文构建失败")
			continue
		}

		ctl.hub.BroadcastStarting(userID)

		aiResponse, err := ctl.aiService.SendMessage(conv, content, history, func(chunk string) {
			ctl.hub.SendStreamChunk(userID, chunk)
		})
		if err != nil {
			log.Printf("AI SendMessage error: %v", err)
			ctl.hub.SendError(userID, "AI 响应失败: "+err.Error())
			continue
		}

		aiMessage := &model.Message{
			ConversationID: conv.ID,
			Role:           "assistant",
			Content:        aiResponse,
		}
		if err := ctl.msgRepo.Create(aiMessage); err != nil {
			log.Printf("Save AI message error: %v", err)
		}

		ctl.hub.SendComplete(userID, aiResponse)
	}
}

func (ctl *ChatController) writePump(client *service.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
