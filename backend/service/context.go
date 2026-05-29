package service

import (
	"encoding/json"
	"fmt"
	"time"

	"rain-yi-backend/config"
	"rain-yi-backend/model"
	"rain-yi-backend/repository"

	"github.com/redis/go-redis/v9"
)

const MaxContextLength = 20

type ContextManager struct {
	msgRepo *repository.MessageRepository
}

func NewContextManager(msgRepo *repository.MessageRepository) *ContextManager {
	return &ContextManager{msgRepo: msgRepo}
}

func (cm *ContextManager) BuildContext(convID int64) ([]model.Message, error) {
	if config.RDB != nil {
		messages, err := cm.getFromRedis(convID)
		if err == nil && len(messages) > 0 {
			if len(messages) > MaxContextLength {
				messages = messages[len(messages)-MaxContextLength:]
			}
			return messages, nil
		}
	}

	messages, err := cm.msgRepo.GetRecentMessages(convID, MaxContextLength)
	if err != nil {
		return nil, err
	}

	if config.RDB != nil && len(messages) > 0 {
		cm.saveToRedis(convID, messages)
	}

	return messages, nil
}

func (cm *ContextManager) AppendToContext(convID int64, msg *model.Message) {
	if config.RDB == nil {
		return
	}

	key := fmt.Sprintf("chat:context:%d", convID)
	data, _ := json.Marshal(msg)

	config.RDB.RPush(config.RedisCtx, key, data)
	config.RDB.LTrim(config.RedisCtx, key, -50, -1)
	config.RDB.Expire(config.RedisCtx, key, 24*time.Hour)
}

func (cm *ContextManager) getFromRedis(convID int64) ([]model.Message, error) {
	key := fmt.Sprintf("chat:context:%d", convID)
	vals, err := config.RDB.LRange(config.RedisCtx, key, 0, -1).Result()
	if err != nil || len(vals) == 0 {
		return nil, redis.Nil
	}

	messages := make([]model.Message, 0, len(vals))
	for _, v := range vals {
		var msg model.Message
		if err := json.Unmarshal([]byte(v), &msg); err == nil {
			messages = append(messages, msg)
		}
	}

	config.RDB.Expire(config.RedisCtx, key, 24*time.Hour)
	return messages, nil
}

func (cm *ContextManager) saveToRedis(convID int64, messages []model.Message) {
	key := fmt.Sprintf("chat:context:%d", convID)

	pipe := config.RDB.Pipeline()
	pipe.Del(config.RedisCtx, key)

	for _, msg := range messages {
		data, _ := json.Marshal(msg)
		pipe.RPush(config.RedisCtx, key, data)
	}

	pipe.LTrim(config.RedisCtx, key, -50, -1)
	pipe.Expire(config.RedisCtx, key, 24*time.Hour)
	pipe.Exec(config.RedisCtx)
}

func (cm *ContextManager) TrimContext(messages []model.Message) []model.Message {
	if len(messages) > MaxContextLength {
		start := len(messages) - MaxContextLength
		return messages[start:]
	}
	return messages
}

func (cm *ContextManager) ResetContext(convID int64) error {
	if config.RDB != nil {
		config.RDB.Del(config.RedisCtx, fmt.Sprintf("chat:context:%d", convID))
	}
	return cm.msgRepo.SoftDeleteByConversationID(convID)
}
