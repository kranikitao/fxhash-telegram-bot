package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	DeliveryItemTypeByArtist = "by_artist"
	DeliveryItemTypeFree     = "free"
	NullChatID               = -1
)

type DeliveryItem struct {
	gorm.Model
	ID             uint64    `gorm:"column:id"`
	Type           string    `gorm:"column:type;index:uidx_type_chat_id_generative_id,unique"`
	ChatID         int64     `gorm:"column:chat_id;index:uidx_type_chat_id_generative_id,unique"`
	GenerativeId   int64     `gorm:"column:generative_id;index:uidx_type_chat_id_generative_id,unique"`
	GenerativeSlug string    `gorm:"column:generative_slug"`
	IsSent         bool      `gorm:"column:is_sent;index:idx_is_sent"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
	Url            string    `gorm:"column:url"`
}

func (m DeliveryItem) TableName() string {
	return "delivery_items"
}

// func (m DeliveryItem) Is() string {
// 	return "delivery_items"
// }
