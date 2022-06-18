package messagesender

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm/model"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

type Sender struct {
	logger            *zap.Logger
	bot               *tgbotapi.BotAPI
	deliveryItemStore *orm.DeliveryItemStore
	subscriberStore   *orm.SubscriberStore
}

func New(logger *zap.Logger, bot *tgbotapi.BotAPI, gorm *gorm.DB) *Sender {
	return &Sender{
		logger:            logger,
		bot:               bot,
		deliveryItemStore: orm.GetDeliveryItemStore(gorm),
		subscriberStore:   orm.GetSubscriberStore(gorm),
	}
}

func (s *Sender) Start() {
	done := make(chan bool)
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-done:
			ticker.Stop()
		case <-ticker.C:
			deliveryItems, err := s.deliveryItemStore.FindNotSent()
			if err != nil {
				s.logger.Error("can't get delivery items",
					zap.Error(err),
					errors.ErrorTraceLogField(err),
				)
				continue
			}
			subscribers, err := s.subscriberStore.FindSubscribed()
			if err != nil && err.Type != orm.ErrNotFound {
				s.logger.Error(
					"can't get active subscribers",
					zap.Error(err),
					errors.ErrorTraceLogField(err),
				)
			}
			for _, item := range deliveryItems {
				if item.ChatID == model.NullChatID {
					for _, subscriber := range subscribers {
						s.sendMessage(subscriber.ChatID, item)
					}
				} else {
					s.sendMessage(item.ChatID, item)
				}
				item.IsSent = true
				if err := s.deliveryItemStore.Update(item); err != nil {
					s.logger.Error(
						"can't update item",
						zap.Any("item", item),
						zap.Error(err),
						errors.ErrorTraceLogField(err),
					)
				}
			}
		}
	}
}

func (s *Sender) sendMessage(ChatID int64, item *model.DeliveryItem) {
	messageText := "A new generative art has appeared on the fxhash: " + item.Url
	message := tgbotapi.NewMessage(ChatID, messageText)
	_, err := s.bot.Send(message)
	if err != nil {
		err := errors.Wrap(err, "")
		s.logger.Error(
			"can't send message with generative",
			zap.Any("item", item),
			zap.Int64("chatID", ChatID),
			zap.Error(err),
			errors.ErrorTraceLogField(err),
		)
	}
}
