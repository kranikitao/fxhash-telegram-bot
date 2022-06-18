package model

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	ID        uint64    `gorm:"column:id"`
	ChatID    int64     `gorm:"column:chat_id"`
	EventCode string    `gorm:"column:event_code"`
	EventData string    `gorm:"column:event_data"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (m Event) TableName() string {
	return "events"
}
