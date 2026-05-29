package repository

import (
	"rain-yi-backend/config"
	"rain-yi-backend/model"
)

type ConversationRepository struct{}

func NewConversationRepository() *ConversationRepository {
	return &ConversationRepository{}
}

func (r *ConversationRepository) Create(conv *model.Conversation) error {
	return config.DB.Create(conv).Error
}

func (r *ConversationRepository) FindByUserID(userID int64) ([]model.Conversation, error) {
	var conversations []model.Conversation
	err := config.DB.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&conversations).Error
	return conversations, err
}

func (r *ConversationRepository) FindByID(id int64) (*model.Conversation, error) {
	var conv model.Conversation
	err := config.DB.First(&conv, id).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepository) Update(conv *model.Conversation) error {
	return config.DB.Save(conv).Error
}

func (r *ConversationRepository) UpdateAIAvatar(convID int64, avatarURL string) error {
	return config.DB.Model(&model.Conversation{}).Where("id = ?", convID).Update("ai_avatar", avatarURL).Error
}

func (r *ConversationRepository) FindOrCreateDefault(userID int64) (*model.Conversation, error) {
	var conv model.Conversation
	err := config.DB.Where("user_id = ?", userID).
		Order("updated_at DESC").
		First(&conv).Error

	if err == nil {
		return &conv, nil
	}

	conv = model.Conversation{
		UserID:     userID,
		Title:      "情感陪伴",
		AINickname: "RainYi",
		AIAvatar:   "/static/default-avatar.svg",
	}

	if err := config.DB.Create(&conv).Error; err != nil {
		return nil, err
	}

	return &conv, nil
}
