package sender

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

type Sender struct {
	logger            *zap.Logger
	bot               *tgbotapi.BotAPI
	deliveryItemStore *orm.DeliveryItemStore
}

func New(logger *zap.Logger, bot *tgbotapi.BotAPI, gorm *gorm.DB) *Sender {
	return &Sender{
		logger:            logger,
		bot:               bot,
		deliveryItemStore: orm.GetDeliveryItemStore(gorm),
	}
}

func (s *Sender) Start() {
	done := make(chan bool)
	ticker := time.NewTicker(60 * time.Second)
	go func() {
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
				for _, item := range deliveryItems {
					messageText := "A new token has appeared on the fxhash: " + item.Url
					message := tgbotapi.NewMessage(item.ChatID, messageText)
					_, err := s.bot.Send(message)
					if err != nil {
						err := errors.Wrap(err, "")
						s.logger.Error(
							"can't send message with generative",
							zap.Any("item", item),
							zap.Error(err),
							errors.ErrorTraceLogField(err),
						)
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
	}()
}
