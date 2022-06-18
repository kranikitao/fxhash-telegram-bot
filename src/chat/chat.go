package chat

import (
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
	"github.com/kranikitao/fxhash-telegram-bot/src/fxhash"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	ChatErrorUnexpected = "Unexpected error. I am working on it."
)

type Chat struct {
	bot                     *tgbotapi.BotAPI
	subscriberStore         *orm.SubscriberStore
	fxHash                  *fxhash.FxHash
	logger                  *zap.Logger
	eventStore              *orm.EventStore
	artistSubscriptionStore *orm.ArtistSubscriptionStore
}

func New(bot *tgbotapi.BotAPI, logger *zap.Logger, gorm *gorm.DB) *Chat {
	return &Chat{
		bot:                     bot,
		fxHash:                  fxhash.New(),
		logger:                  logger,
		eventStore:              orm.GetEventStore(gorm),
		subscriberStore:         orm.GetSubscriberStore(gorm),
		artistSubscriptionStore: orm.GetArtistSubscriptionStore(gorm),
	}
}

func (c *Chat) Start() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := c.bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		var currentMessage *tgbotapi.Message
		isCallbackQuery := false
		if update.Message != nil {
			currentMessage = update.Message
		} else if update.CallbackQuery != nil {
			isCallbackQuery = true
			currentMessage = update.CallbackQuery.Message
		} else {
			continue
		}
		subscriber, err := c.registerSubscriberIfNotExists(currentMessage)
		if err != nil {
			c.logger.Error(
				"Can't register subscriber",
				zap.Any("Chat", update.Message.Chat),
				zap.Any("From", update.Message.From),
				zap.Error(err),
				errors.ErrorTraceLogField(err),
			)
			c.sendTextMessage(currentMessage.Chat.ID, ChatErrorUnexpected)
			continue
		}

		if !isCallbackQuery {
			c.PushEvent(subscriber.ChatID, "chat", update.Message.Text)
			if currentMessage.IsCommand() {
				c.handleCommmands(update.Message.Command(), currentMessage.CommandArguments(), subscriber)
			} else {
				if subscriber.State == CommandSubscribeArtist {
					c.subscribeToArtist(currentMessage.Text, subscriber)
				}
			}
		} else {
			c.PushEvent(subscriber.ChatID, "callback", update.CallbackQuery.Data)
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := c.bot.Request(callback); err != nil {
				panic(err)
			}

			data := update.CallbackQuery.Data
			command, arguments := c.parseCommandAndArguments(data)
			switch command {
			case CommandCancel:
				if err := c.updateState(subscriber, ""); err != nil {
					c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
				} else {
					deleteRequest := tgbotapi.NewDeleteMessage(subscriber.ChatID, update.CallbackQuery.Message.MessageID)
					if _, err := c.bot.Request(deleteRequest); err != nil {
						c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
					}
				}
			case CommandUnsubscribeFree:
				subscriber.Subscribed = false
				if err := c.subscriberStore.Update(subscriber); err != nil {
					c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
				} else {
					deleteRequest := tgbotapi.NewDeleteMessage(subscriber.ChatID, update.CallbackQuery.Message.MessageID)
					if _, err := c.bot.Request(deleteRequest); err != nil {
						c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
					} else {
						c.showUnsubscribeWindow(subscriber)
					}
				}
			case CommandUnsubscribe:
				if arguments != "" {
					textToSend := ""
					artistSubscripton, err := c.artistSubscriptionStore.FindByChatIDAndFxHashArtistName(subscriber.ChatID, arguments)
					if err != nil {
						if err.Type == orm.ErrNotFound {
							textToSend = "Subscription not found."
						} else {
							textToSend = ChatErrorUnexpected
							c.logger.Error(
								"can't get subscription",
								zap.Int64("chatId", subscriber.ChatID),
								zap.String("message", arguments),
								zap.Error(err),
								errors.ErrorTraceLogField(err),
							)
						}
					} else {
						artistSubscripton.IsActive = false
						if err := c.artistSubscriptionStore.Update(artistSubscripton); err != nil {
							textToSend = ChatErrorUnexpected
							c.logger.Error(
								"can't update subscription",
								zap.Int64("chatId", subscriber.ChatID),
								zap.String("message", arguments),
								zap.Error(err),
								errors.ErrorTraceLogField(err),
							)
						}
					}
					if err := c.updateState(subscriber, ""); err != nil {
						textToSend = ChatErrorUnexpected
					}
					if textToSend == "" {
						deleteRequest := tgbotapi.NewDeleteMessage(subscriber.ChatID, update.CallbackQuery.Message.MessageID)
						if _, err := c.bot.Request(deleteRequest); err != nil {
							c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
						} else {
							c.showUnsubscribeWindow(subscriber)
						}
					} else {
						message := tgbotapi.NewMessage(subscriber.ChatID, textToSend)
						message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						_, errr := c.bot.Send(message)
						if errr != nil {
							errr := errors.Wrap(err, "")
							c.logger.Error(
								"can't send message with keyboard",
								zap.Any("message", message),
								zap.Error(errr),
								errors.ErrorTraceLogField(errr),
							)
						}
					}
				} else {
					c.showUnsubscribeWindow(subscriber)
				}
			}
		}
	}
}

func (c *Chat) showUnsubscribeWindow(subscriber *model.Subscriber) {
	subscribtions, err := c.artistSubscriptionStore.FindActiveByChatId(subscriber.ChatID)
	if err != nil {
		c.logger.Error(
			"can't get subscriptions",
			zap.Int64("chatId", subscriber.ChatID),
			zap.Error(err),
			errors.ErrorTraceLogField(err),
		)
		c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
	}
	if len(subscribtions) > 0 || subscriber.Subscribed {
		var buttons [][]tgbotapi.InlineKeyboardButton
		for _, subscription := range subscribtions {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					subscription.FxHashArtistName,
					"/"+CommandUnsubscribe+" "+subscription.FxHashArtistName,
				)),
			)
		}
		if subscriber.Subscribed {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"Unsubscribe free generatives",
					"/"+CommandUnsubscribeFree,
				)),
			)
		}
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Cancel", "/"+CommandCancel)))
		var keyboard = tgbotapi.NewInlineKeyboardMarkup(buttons...)
		message := tgbotapi.NewMessage(subscriber.ChatID, "Select the artist you want to unsubscribe")
		message.ReplyMarkup = keyboard
		_, errr := c.bot.Send(message)
		if errr != nil {
			errr := errors.Wrap(errr, "")
			c.logger.Error(
				"can't send message with keyboard",
				zap.Any("message", message),
				zap.Error(errr),
				errors.ErrorTraceLogField(errr),
			)
		}
	} else {
		c.sendTextMessage(subscriber.ChatID, "There are no subscriptions.")
	}
}

func (*Chat) parseCommandAndArguments(data string) (string, string) {
	splittedData := strings.Split(data, " ")
	command := data
	arguments := ""
	if len(splittedData) > 1 {
		command = splittedData[0]
		arguments = data[len(command)+1:]
	}
	if !strings.HasPrefix(command, "/") {
		command = ""
	} else {
		command = command[1:]
	}

	return command, arguments
}

func (c *Chat) subscribeToArtist(textRecieved string, subscriber *model.Subscriber) {
	textRecieved = strings.ReplaceAll(textRecieved, "%20", " ")
	splitedUrl := strings.Split(textRecieved, "/")
	found := false
	fxHashUserName := ""
	for _, partOfUrl := range splitedUrl {
		if found {
			fxHashUserName = partOfUrl
		}
		if partOfUrl == "u" {
			found = true
		}
	}
	if fxHashUserName == "" {
		c.sendTextMessage(subscriber.ChatID, "Unrecognized url, please try again.")
	}

	user, err := c.fxHash.GetFxHashUser(fxHashUserName)
	if err != nil {
		if err.Type == fxhash.ErrTypeUserNotFound {
			c.sendTextMessage(subscriber.ChatID, fmt.Sprintf("FxHash user %s not found.", fxHashUserName))
		} else {
			c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
		}
		return
	}

	subscription, err := c.artistSubscriptionStore.FindByChatIDAndFxHashArtistName(subscriber.ChatID, user.Name)
	if err != nil {
		if err.Type == orm.ErrNotFound {
			subscription = &model.ArtistSubscribtion{
				ChatID:           subscriber.ChatID,
				FxHashArtistName: user.Name,
				FxHashArtistID:   user.Id,
				IsActive:         true,
			}
			if err := c.artistSubscriptionStore.Create(subscription); err != nil {
				c.logger.Error(
					"can't add subscription",
					zap.Any("subscription", subscription),
					zap.Error(err),
					errors.ErrorTraceLogField(err),
				)
				c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
			}
		} else {
			c.logger.Error(
				"can't get subscription",
				zap.Any("subscription", subscription),
				zap.Error(err),
				errors.ErrorTraceLogField(err),
			)
			c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
			return
		}
	} else {
		subscription.IsActive = true
		if err := c.artistSubscriptionStore.Update(subscription); err != nil {
			c.logger.Error(
				"can't update subscription",
				zap.Any("subscription", subscription),
				zap.Error(err),
				errors.ErrorTraceLogField(err),
			)
			c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
		}
	}

	if err := c.updateState(subscriber, ""); err != nil {
		c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
	}

	c.sendTextMessage(subscriber.ChatID, fmt.Sprintf("You was subscribed to %s.", fxHashUserName))
}

func (c *Chat) updateState(subscriber *model.Subscriber, state string) *errors.Error {
	subscriber.State = state
	if err := c.subscriberStore.Update(subscriber); err != nil {
		c.logger.Error(
			"can't update state",
			zap.Any("subscriber", subscriber),
			zap.Error(err),
			errors.ErrorTraceLogField(err),
		)
		return err
	}

	return nil
}

func (c *Chat) handleCommmands(command string, arguments string, subscriber *model.Subscriber) {
	switch command {
	case CommandStart:
		c.answerStart(subscriber.ChatID, subscriber.Username)
	case CommandSubscribeFree:
		subscriber.Subscribed = true
		if err := c.subscriberStore.Update(subscriber); err != nil {
			c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
		} else {
			c.sendTextMessage(subscriber.ChatID, "You was subscribed to zero cost minting generatives")
		}
	case CommandUnsubscribe:
		c.showUnsubscribeWindow(subscriber)
	case CommandSubscribeArtist:
		if err := c.updateState(subscriber, CommandSubscribeArtist); err != nil {
			c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
		} else {
			c.sendTextMessage(subscriber.ChatID, "Type link to artist on fx hash \n(ex: https://www.fxhash.xyz/u/kranikitao)")
		}
	case CommandCancel:
		if err := c.updateState(subscriber, ""); err != nil {
			c.sendTextMessage(subscriber.ChatID, ChatErrorUnexpected)
		} else {
			c.sendTextMessage(subscriber.ChatID, "Operation was canceled")
		}
	default:
	}
}

func (c *Chat) registerSubscriberIfNotExists(message *tgbotapi.Message) (*model.Subscriber, *errors.Error) {
	subscriber, err := c.subscriberStore.FindByChatID(message.Chat.ID)
	if err != nil {
		if err.Type == orm.ErrNotFound {
			buf, err := json.Marshal(message.From)
			rawUser := ""
			if err != nil {
				c.logger.Error(
					"can't parse From",
					zap.Any("From", message.From),
					zap.Error(err),
				)
			} else {
				rawUser = string(buf)
			}

			subscriber = &model.Subscriber{
				ChatID:     message.Chat.ID,
				Username:   message.From.UserName,
				Subscribed: false,
				RawUser:    rawUser,
				State:      "",
			}

			if err := c.subscriberStore.Create(subscriber); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return subscriber, nil
}

func (c *Chat) answerStart(chatId int64, firstName string) *errors.Error {
	welcomeText := "Hello, " + firstName + "!\n" +

		"I can help you to be first minter on fxhash.xyz.\n" +
		"So, first of all you should subscribe. \n\n" +
		"There are two types of subscription:\n" +
		"/subscribeartist - subscription to new generatives of your favorite artist\n" +
		"/subscribefree - subscription to zero cost minting generatives\n\n" +
		"Type /unsubscribe to manage subscriptions:\n\n" +
		"Author @kranikitao\n"

	return c.sendTextMessage(chatId, welcomeText)
}

func (c *Chat) sendTextMessage(chatId int64, text string) *errors.Error {
	message := tgbotapi.NewMessage(chatId, text)
	_, err := c.bot.Send(message)
	if err != nil {
		err := errors.Wrap(err, "")
		c.logger.Error(
			"can't send message",
			zap.Int64("ChatId", chatId),
			zap.String("text", text),
			zap.Error(err),
			errors.ErrorTraceLogField(err),
		)
		return err
	}

	return nil
}

func (c *Chat) PushEvent(chatId int64, eventCode string, eventData string) {
	if err := c.eventStore.Push(chatId, eventCode, eventData); err != nil {
		c.logger.Error("can't push event",
			zap.Error(err),
			errors.ErrorTraceLogField(err),
		)
	}
}
