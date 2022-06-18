package model

import (
	"time"

	"gorm.io/gorm"
)

type ArtistSubscribtion struct {
	gorm.Model
	ID               uint64    `gorm:"column:id"`
	FxHashArtistName string    `gorm:"column:fx_hash_artist_name"`
	FxHashArtistID   string    `gorm:"column:fx_hash_artist_id"`
	ChatID           int64     `gorm:"column:chat_id"`
	IsActive         bool      `gorm:"column:is_active"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (m ArtistSubscribtion) TableName() string {
	return "artist_subscriptions"
}
