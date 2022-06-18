package orm

import (
	"time"

	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm/model"
	"gorm.io/gorm"
)

type DeliveryItemStore struct {
	gorm *gorm.DB
}

func GetDeliveryItemStore(gorm *gorm.DB) *DeliveryItemStore {
	return &DeliveryItemStore{
		gorm: gorm,
	}
}

func (s *DeliveryItemStore) Create(m *model.DeliveryItem) *errors.Error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	result := s.gorm.Create(&m)
	if result.Error != nil {
		return errors.Wrap(result.Error, "")
	}

	return nil
}

func (s *DeliveryItemStore) Update(m *model.DeliveryItem) *errors.Error {
	m.UpdatedAt = time.Now()
	result := s.gorm.Save(&m)
	if result.Error != nil {
		return errors.Wrap(result.Error, "")
	}

	return nil
}

func (s *DeliveryItemStore) FindByTypeAndChatIdAndGenerativeId(Type string, chatId int64, generativeId int64) (*model.DeliveryItem, *errors.Error) {
	m := &model.DeliveryItem{}
	result := s.gorm.Where("type = ? AND chat_id = ? AND generative_id = ?", Type, chatId, generativeId).First(&m)

	return wrapSingleResult(m, result.Error)
}

func (s *DeliveryItemStore) FindNotSent() ([]*model.DeliveryItem, *errors.Error) {
	var m []*model.DeliveryItem
	result := s.gorm.Where("is_sent = false").Find(&m)

	return wrapListResult(m, result.Error)
}
