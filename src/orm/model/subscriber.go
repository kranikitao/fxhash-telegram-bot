package model

import (
	"time"

	"gorm.io/gorm"
)

type Subscriber struct {
	gorm.Model
	ID         uint64    `gorm:"column:id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
	ChatID     int64     `gorm:"column:chat_id;index:uidx_chat_id,unique"`
	Username   string    `gorm:"column:username"`
	Subscribed bool      `gorm:"column:subscribed"`
	State      string    `gorm:"column:state"`
	RawUser    string    `gorm:"column:raw_user"`
}

func (m Subscriber) TableName() string {
	return "subscribers"
}
