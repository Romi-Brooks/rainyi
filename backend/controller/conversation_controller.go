package controller

import (
	"net/http"
	"strconv"

	"rain-yi-backend/model"
	"rain-yi-backend/repository"
	"rain-yi-backend/service"
	"rain-yi-backend/utils"

	"github.com/gin-gonic/gin"
)

type ConversationController struct {
	convRepo       *repository.ConversationRepository
	msgRepo        *repository.MessageRepository
	contextManager *service.ContextManager
}

func NewConversationController(
	convRepo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	contextManager *service.ContextManager,
) *ConversationController {
	return &ConversationController{
		convRepo:       convRepo,
		msgRepo:        msgRepo,
		contextManager: contextManager,
	}
}

func (ctl *ConversationController) GetConversations(c *gin.Context) {
	userID := c.GetInt64("user_id")

	conversations, err := ctl.convRepo.FindByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会话列表失败"})
		return
	}

	if len(conversations) == 0 {
		conv, err := ctl.convRepo.FindOrCreateDefault(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建默认会话失败"})
			return
		}
		conversations = append(conversations, *conv)
	}

	for i := range conversations {
		lastMsg, err := ctl.msgRepo.GetLastMessage(conversations[i].ID)
		if err == nil && lastMsg != nil {
			conversations[i].LastMessage = &model.LastMessage{
				Content:   lastMsg.Content,
				Role:      lastMsg.Role,
				CreatedAt: lastMsg.CreatedAt,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
	})
}

func (ctl *ConversationController) GetMessages(c *gin.Context) {
	convIDStr := c.Param("id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID无效"})
		return
	}

	conv, err := ctl.convRepo.FindByID(convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	userID := c.GetInt64("user_id")
	if conv.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该会话"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	messages, err := ctl.msgRepo.FindByConversationID(convID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息失败"})
		return
	}

	total, _ := ctl.msgRepo.CountByConversationID(convID)

	c.JSON(http.StatusOK, gin.H{
		"messages":  messages,
		"total":     total,
		"conversation": conv,
	})
}

func (ctl *ConversationController) ClearMessages(c *gin.Context) {
	convIDStr := c.Param("id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID无效"})
		return
	}

	conv, err := ctl.convRepo.FindByID(convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	userID := c.GetInt64("user_id")
	if conv.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作该会话"})
		return
	}

	if err := ctl.contextManager.ResetContext(convID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "聊天记录已清空"})
}

func (ctl *ConversationController) UpdateConfig(c *gin.Context) {
	convIDStr := c.Param("id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID无效"})
		return
	}

	conv, err := ctl.convRepo.FindByID(convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	userID := c.GetInt64("user_id")
	if conv.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作该会话"})
		return
	}

	var req struct {
		AINickname *string `json:"ai_nickname"`
		AIAvatar   *string `json:"ai_avatar"`
		Title      *string `json:"title"`
		PersonaID  *int64  `json:"persona_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.PersonaID != nil {
		conv.PersonaID = req.PersonaID
	}
	if req.AINickname != nil {
		conv.AINickname = utils.SanitizeInput(*req.AINickname)
	}
	if req.AIAvatar != nil {
		conv.AIAvatar = *req.AIAvatar
	}
	if req.Title != nil {
		conv.Title = utils.SanitizeInput(*req.Title)
	}

	if err := ctl.convRepo.Update(conv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "更新成功",
		"conversation": conv,
	})
}

func (ctl *ConversationController) CreateConversation(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Title = "情感陪伴"
	}

	conv := &model.Conversation{
		UserID:     userID,
		Title:      utils.SanitizeInput(req.Title),
		AINickname: "RainYi",
		AIAvatar:   "/static/default-avatar.svg",
	}

	if err := ctl.convRepo.Create(conv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "创建成功",
		"conversation": conv,
	})
}
