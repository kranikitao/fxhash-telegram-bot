package orm

import (
	"time"

	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm/model"
	"gorm.io/gorm"
)

type SubscriberStore struct {
	gorm *gorm.DB
}

func GetSubscriberStore(gorm *gorm.DB) *SubscriberStore {
	return &SubscriberStore{
		gorm: gorm,
	}
}

func (s *SubscriberStore) Create(m *model.Subscriber) *errors.Error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	result := s.gorm.Create(&m)
	if result.Error != nil {
		return errors.Wrap(result.Error, "")
	}

	return nil
}

func (s *SubscriberStore) Update(m *model.Subscriber) *errors.Error {
	m.UpdatedAt = time.Now()
	result := s.gorm.Save(&m)
	if result.Error != nil {
		return errors.Wrap(result.Error, "")
	}

	return nil
}

func (s *SubscriberStore) FindByChatID(chatID int64) (*model.Subscriber, *errors.Error) {
	m := &model.Subscriber{}
	result := s.gorm.Where("chat_id = ?", chatID).First(&m)

	return wrapSingleResult(m, result.Error)
}
