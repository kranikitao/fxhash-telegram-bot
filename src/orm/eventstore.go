package orm

import (
	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm/model"
	"gorm.io/gorm"
)

type EventStore struct {
	gorm *gorm.DB
}

func GetEventStore(gorm *gorm.DB) *EventStore {
	return &EventStore{
		gorm: gorm,
	}
}

func (s *EventStore) Push(chatId int64, eventCode string, eventData string) *errors.Error {
	e := &model.Event{
		ChatID:    chatId,
		EventCode: eventCode,
		EventData: eventData,
	}
	result := s.gorm.Create(&e)
	if result.Error != nil {
		return errors.Wrap(result.Error, "")
	}

	return nil
}
