// Package repo provides GORM repositories for messages
package repo

import (
	"gorm.io/gorm"
)

type Message struct {
	ID      uint   `gorm:"primaryKey"`
	To      string
	Content string
	Status  string
}

func SaveMessage(db *gorm.DB, msg *Message) error {
	return db.Create(msg).Error
}
