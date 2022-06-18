package orm

import (
	"time"

	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm/model"
	"gorm.io/gorm"
)

type ArtistSubscriptionStore struct {
	gorm *gorm.DB
}

func GetArtistSubscriptionStore(gorm *gorm.DB) *ArtistSubscriptionStore {
	return &ArtistSubscriptionStore{
		gorm: gorm,
	}
}

func (s *ArtistSubscriptionStore) Create(m *model.ArtistSubscribtion) *errors.Error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	result := s.gorm.Create(&m)
	if result.Error != nil {
		return errors.Wrap(result.Error, "")
	}

	return nil
}

func (s *ArtistSubscriptionStore) Update(m *model.ArtistSubscribtion) *errors.Error {
	m.UpdatedAt = time.Now()
	result := s.gorm.Save(&m)
	if result.Error != nil {
		return errors.Wrap(result.Error, "")
	}

	return nil
}

func (s *ArtistSubscriptionStore) FindByChatIDAndFxHashArtistName(chatID int64, FxHashArtistName string) (*model.ArtistSubscribtion, *errors.Error) {
	var m *model.ArtistSubscribtion
	result := s.gorm.Where("chat_id = ? AND fx_hash_artist_name = ?", chatID, FxHashArtistName).First(&m)

	return wrapSingleResult(m, result.Error)
}

func (s *ArtistSubscriptionStore) FindActiveByChatId(chatID int64) ([]*model.ArtistSubscribtion, *errors.Error) {
	var m []*model.ArtistSubscribtion
	result := s.gorm.Where("chat_id = ? AND is_active = true", chatID).Find(&m)

	return wrapListResult(m, result.Error)
}

func (s *ArtistSubscriptionStore) FindActiveByFxHashArtistIds(fxHashArtistIds []string) ([]*model.ArtistSubscribtion, *errors.Error) {
	var m []*model.ArtistSubscribtion
	result := s.gorm.Where("fx_hash_artist_id IN ? AND is_active = true", fxHashArtistIds).Find(&m)

	return wrapListResult(m, result.Error)
}
