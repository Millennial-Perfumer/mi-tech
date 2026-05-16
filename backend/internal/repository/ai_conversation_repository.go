package repository

import (
	"encoding/json"
	"mi-tech/internal/entity"
	"time"

	"gorm.io/gorm"
)

type AIConversationRepository interface {
	CreateConversation(userID int64, title string) (*entity.AIConversation, error)
	ListConversations(userID int64, limit int) ([]entity.AIConversation, error)
	GetConversation(id, userID int64) (*entity.AIConversation, error)
	DeleteConversation(id, userID int64) error
	UpdateTitle(id int64, title string) error

	AddMessage(conversationID int64, role, content string, metadata *json.RawMessage) error
	GetMessages(conversationID int64) ([]entity.AIMessage, error)
}

type gormAIConversationRepository struct {
	db *gorm.DB
}

func NewAIConversationRepository(db *gorm.DB) AIConversationRepository {
	return &gormAIConversationRepository{db: db}
}

func (r *gormAIConversationRepository) CreateConversation(userID int64, title string) (*entity.AIConversation, error) {
	conv := &entity.AIConversation{
		UserID: userID,
		Title:  title,
	}
	err := r.db.Create(conv).Error
	return conv, err
}

func (r *gormAIConversationRepository) ListConversations(userID int64, limit int) ([]entity.AIConversation, error) {
	var convs []entity.AIConversation
	err := r.db.Where("user_id = ?", userID).Order("updated_at DESC").Limit(limit).Find(&convs).Error
	return convs, err
}

func (r *gormAIConversationRepository) GetConversation(id, userID int64) (*entity.AIConversation, error) {
	var conv entity.AIConversation
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&conv).Error
	return &conv, err
}

func (r *gormAIConversationRepository) DeleteConversation(id, userID int64) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.AIConversation{}).Error
}

func (r *gormAIConversationRepository) UpdateTitle(id int64, title string) error {
	return r.db.Model(&entity.AIConversation{}).Where("id = ?", id).Update("title", title).Error
}

func (r *gormAIConversationRepository) AddMessage(conversationID int64, role, content string, metadata *json.RawMessage) error {
	msg := &entity.AIMessage{
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		Metadata:       metadata,
	}
	
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}
		// Update conversation timestamp
		return tx.Model(&entity.AIConversation{}).Where("id = ?", conversationID).Update("updated_at", time.Now()).Error
	})
	
	return err
}

func (r *gormAIConversationRepository) GetMessages(conversationID int64) ([]entity.AIMessage, error) {
	var msgs []entity.AIMessage
	err := r.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}
