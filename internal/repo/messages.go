// Package repo provides GORM-backed persistence for messages.
package repo

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Message is the persisted record of a single send request.
type Message struct {
	ID         uint      `gorm:"primaryKey"`
	ExternalID string    `gorm:"column:external_id;uniqueIndex"`
	Channel    string    `gorm:"column:channel"`
	Recipient  string    `gorm:"column:recipient"`
	Content    string    `gorm:"column:content"`
	Status     string    `gorm:"column:status"`
	Provider   string    `gorm:"column:provider"`
	ProviderID string    `gorm:"column:provider_id"`
	Error      string    `gorm:"column:error"`
	Attempts   int       `gorm:"column:attempts"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (Message) TableName() string { return "messages" }

// Statuses a message can be in.
const (
	StatusQueued     = "queued"
	StatusSent       = "sent"
	StatusFailed     = "failed"
	StatusDeadLetter = "dead_letter"
)

// MessageRepository persists and retrieves Message records.
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository builds a MessageRepository backed by db.
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create inserts a new message row.
func (r *MessageRepository) Create(ctx context.Context, msg *Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// GetByExternalID looks up a message by its public-facing ID.
func (r *MessageRepository) GetByExternalID(ctx context.Context, externalID string) (*Message, error) {
	var msg Message
	if err := r.db.WithContext(ctx).Where("external_id = ?", externalID).First(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

// UpdateResult records the outcome of a send attempt.
func (r *MessageRepository) UpdateResult(ctx context.Context, externalID, status, providerID, sendErr string, attempts int) error {
	return r.db.WithContext(ctx).Model(&Message{}).Where("external_id = ?", externalID).Updates(map[string]any{
		"status":      status,
		"provider_id": providerID,
		"error":       sendErr,
		"attempts":    attempts,
	}).Error
}
