package artcollector

import (
	"time"

	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
	"github.com/kranikitao/fxhash-telegram-bot/src/fxhash"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm"
	"github.com/kranikitao/fxhash-telegram-bot/src/orm/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ArtCollector struct {
	logger                  *zap.Logger
	fxhash                  *fxhash.FxHash
	deliveryItemStore       *orm.DeliveryItemStore
	artistSubscriptionStore *orm.ArtistSubscriptionStore
}

func New(logger *zap.Logger, fxhash *fxhash.FxHash, gorm *gorm.DB) *ArtCollector {
	return &ArtCollector{
		logger:                  logger,
		fxhash:                  fxhash,
		deliveryItemStore:       orm.GetDeliveryItemStore(gorm),
		artistSubscriptionStore: orm.GetArtistSubscriptionStore(gorm),
	}
}

func (c *ArtCollector) Collect() {
	done := make(chan bool)
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
			case <-ticker.C:
				tokens, err := c.fxhash.GetLastGeneratives()
				if err != nil {
					c.logger.Error("can't get generatives",
						zap.Error(err),
						errors.ErrorTraceLogField(err),
					)
				}

				tokensByAuthors := map[string][]*fxhash.GenerativeToken{}
				var authorIds []string
				for _, token := range tokens {
					if token.Author.Name != "" {
						authorIds = append(authorIds, token.Author.Id)
						tokensByAuthors[token.Author.Id] = append(tokensByAuthors[token.Author.Id], token)
					} else {
						for _, author := range token.Collaborators {
							if author.Name != "" {
								authorIds = append(authorIds, author.Id)
								tokensByAuthors[author.Id] = append(tokensByAuthors[token.Author.Id], token)
							}
						}
					}
				}
				subscriptions, err := c.artistSubscriptionStore.FindActiveByFxHashArtistIds(authorIds)
				if err != nil {
					c.logger.Error("can't get subscriptions",
						zap.Error(err),
						errors.ErrorTraceLogField(err),
					)
				}
				if len(subscriptions) == 0 {
					continue
				}

				for _, subscription := range subscriptions {
					if tokensByAuthors[subscription.FxHashArtistID] != nil {
						for _, token := range tokensByAuthors[subscription.FxHashArtistID] {
							deliveryItem := &model.DeliveryItem{
								Type:           "by_artist",
								ChatID:         subscription.ChatID,
								GenerativeId:   token.Id,
								GenerativeSlug: token.Slug,
								Url:            "https://www.fxhash.xyz/generative/slug/" + token.Slug,
								IsSent:         false,
							}
							_, err := c.deliveryItemStore.FindByTypeAndChatIdAndGenerativeId("by_artist", deliveryItem.ChatID, deliveryItem.GenerativeId)
							if err != nil {
								if err.Type == orm.ErrNotFound {
									if err := c.deliveryItemStore.Create(deliveryItem); err != nil {
										c.logger.Error("can't add delivery item",
											zap.Any("deliveryItem", deliveryItem),
											zap.Error(err),
											errors.ErrorTraceLogField(err),
										)
									}
								} else {
									c.logger.Error("can't get delivery item",
										zap.Any("deliveryItem", deliveryItem),
										zap.Error(err),
										errors.ErrorTraceLogField(err),
									)
								}
							}
						}
					}
				}
			}
		}
	}()
}
