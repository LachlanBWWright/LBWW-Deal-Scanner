package discord

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"dealscanner/internal/models"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

func downloadFile(url string) (*discordgo.File, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("failed to fetch image: response was nil")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	name := "image.jpg"
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastName := parts[len(parts)-1]
		if idx := strings.Index(lastName, "?"); idx != -1 {
			lastName = lastName[:idx]
		}
		if lastName != "" {
			name = lastName
		}
	}

	return &discordgo.File{
		Name:        name,
		ContentType: resp.Header.Get("Content-Type"),
		Reader:      bytes.NewReader(data),
	}, nil
}

func (b *Bot) SendDeal(ctx context.Context, channelId, channelMessage, dmMessage string, imageUrl *string, queryType, queryId string) error {
	if b.session == nil {
		return fmt.Errorf("bot session is not initialized")
	}

	// 1. Fetch parent queryId and send DMs to subscribed users
	parentQueryId, err := b.getQueryIdByTypeAndKey(ctx, queryType, queryId)
	if err == nil && parentQueryId != "" {
		userQueries, err := b.dbClient.UserQuery.WithContext(ctx).Where(b.dbClient.UserQuery.QueryId.Eq(parentQueryId)).Find()
		if err == nil {
			for _, uq := range userQueries {
				dmChan, err := b.session.UserChannelCreate(uq.UserId)
				if err != nil {
					log.Printf("Failed to create Discord DM channel for user %s: %v", uq.UserId, err)
					continue
				}
				if dmChan == nil {
					log.Printf("Failed to create Discord DM channel for user %s: channel was nil", uq.UserId)
					continue
				}

				unsubscribeActionKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
				if err := b.dbClient.SetAction(ctx, unsubscribeActionKey, models.ActionRegistry{
					ID:        unsubscribeActionKey,
					Type:      "unsubscribe_dm",
					QueryType: &queryType,
					QueryId:   &queryId,
					Timestamp: time.Now().UnixMilli(),
				}); err != nil {
					log.Printf("Failed to register Discord unsubscribe action for user %s: %v", uq.UserId, err)
				}

				dmMsg := &discordgo.MessageSend{
					Content: dmMessage,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Unsubscribe from DM",
									Style:    discordgo.SecondaryButton,
									CustomID: unsubscribeActionKey,
								},
							},
						},
					},
				}
				if imageUrl != nil && *imageUrl != "" {
					if file, err := downloadFile(*imageUrl); err == nil {
						dmMsg.Files = []*discordgo.File{file}
					} else {
						log.Printf("Failed to download image %s for DM: %v", *imageUrl, err)
					}
				}
				if _, err := b.session.ChannelMessageSendComplex(dmChan.ID, dmMsg); err != nil {
					log.Printf("Failed to send Discord DM to user %s: %v", uq.UserId, err)
				}
			}
		} else {
			log.Printf("Failed to load Discord DM subscriptions for query %s: %v", parentQueryId, err)
		}

		// If query is DM-only, do not notify public channel
		parentQuery, err := b.dbClient.SearchQuery.WithContext(ctx).Where(b.dbClient.SearchQuery.ID.Eq(parentQueryId)).First()
		if err == nil && parentQuery.DmOnly {
			return nil
		} else if err != nil {
			log.Printf("Failed to load parent query %s for Discord notification: %v", parentQueryId, err)
		}
	} else if err != nil {
		log.Printf("Failed to resolve parent query for Discord notification %s/%s: %v", queryType, queryId, err)
	}

	// 2. Setup public channel message components
	var buttons []discordgo.MessageComponent

	if queryType != "" && queryId != "" {
		deleteActionKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		if err := b.dbClient.SetAction(ctx, deleteActionKey, models.ActionRegistry{
			ID:        deleteActionKey,
			Type:      "delete",
			QueryType: &queryType,
			QueryId:   &queryId,
			Timestamp: time.Now().UnixMilli(),
		}); err != nil {
			log.Printf("Failed to register Discord delete action for %s/%s: %v", queryType, queryId, err)
		}

		subscribeDMActionKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		if err := b.dbClient.SetAction(ctx, subscribeDMActionKey, models.ActionRegistry{
			ID:        subscribeDMActionKey,
			Type:      "subscribe_dm",
			QueryType: &queryType,
			QueryId:   &queryId,
			Timestamp: time.Now().UnixMilli(),
		}); err != nil {
			log.Printf("Failed to register Discord subscribe action for %s/%s: %v", queryType, queryId, err)
		}

		buttons = append(buttons, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Delete Query",
					Style:    discordgo.DangerButton,
					CustomID: deleteActionKey,
				},
				discordgo.Button{
					Label:    "Subscribe to DM",
					Style:    discordgo.PrimaryButton,
					CustomID: subscribeDMActionKey,
				},
			},
		})
	}

	msgData := &discordgo.MessageSend{
		Content:    channelMessage,
		Components: buttons,
	}

	if imageUrl != nil && *imageUrl != "" {
		if file, err := downloadFile(*imageUrl); err == nil {
			msgData.Files = []*discordgo.File{file}
		} else {
			log.Printf("Failed to download image %s for public channel: %v", *imageUrl, err)
		}
	}

	_, err = b.session.ChannelMessageSendComplex(channelId, msgData)
	return err
}

// SendError publishes an error message to the error channel
func (b *Bot) SendError(channelId, message string) error {
	if b.session == nil {
		return fmt.Errorf("bot session is not initialized")
	}
	_, err := b.session.ChannelMessageSend(channelId, message)
	return err
}

func (b *Bot) getQueryIdByTypeAndKey(ctx context.Context, qType, key string) (string, error) {
	var qId string
	var err error
	switch qType {
	case "salvos":
		var sa *models.Salvos
		sa, err = b.dbClient.Salvos.WithContext(ctx).Where(b.dbClient.Salvos.Name.Eq(key)).First()
		if err == nil {
			qId = sa.QueryId
		}
	case "ebay":
		var eb *models.Ebay
		eb, err = b.dbClient.Ebay.WithContext(ctx).Where(b.dbClient.Ebay.Url.Eq(key)).First()
		if err == nil {
			qId = eb.QueryId
		}
	case "gumtree":
		var gt *models.Gumtree
		gt, err = b.dbClient.Gumtree.WithContext(ctx).Where(b.dbClient.Gumtree.Url.Eq(key)).First()
		if err == nil {
			qId = gt.QueryId
		}
	case "cashConverters":
		var cc *models.CashConvertersFilter
		cc, err = b.dbClient.CashConvertersFilter.WithContext(ctx).Where(b.dbClient.CashConvertersFilter.ID.Eq(key)).First()
		if err == nil {
			qId = cc.QueryId
		}
	case "steamMarket":
		var sm *models.SteamMarket
		sm, err = b.dbClient.SteamMarket.WithContext(ctx).Where(b.dbClient.SteamMarket.Name.Eq(key)).First()
		if err == nil {
			qId = sm.QueryId
		}
	case "csTradeBot":
		var ct *models.CsTradeBot
		ct, err = b.dbClient.CsTradeBot.WithContext(ctx).Where(b.dbClient.CsTradeBot.Name.Eq(key)).First()
		if err == nil {
			qId = ct.QueryId
		}
	case "csMarket":
		var cm *models.CsMarket
		cm, err = b.dbClient.CsMarket.WithContext(ctx).Where(b.dbClient.CsMarket.Url.Eq(key)).First()
		if err == nil {
			qId = cm.QueryId
		}
	default:
		return "", fmt.Errorf("unknown query type: %s", qType)
	}

	if err != nil {
		return "", err
	}
	return qId, nil
}

func (b *Bot) SetStatus(statusText string) {
	if b.session == nil {
		return
	}
	statusText = truncateStatus(statusText)
	usd := discordgo.UpdateStatusData{
		Status: "online",
		Activities: []*discordgo.Activity{
			{
				Name:  "Custom Status",
				Type:  discordgo.ActivityTypeCustom,
				State: statusText,
			},
		},
	}
	b.session.UpdateStatusComplex(usd)
}

const discordStatusCharacterLimit = 128

func truncateStatus(statusText string) string {
	runes := []rune(statusText)
	if len(runes) <= discordStatusCharacterLimit {
		return statusText
	}
	return string(runes[:discordStatusCharacterLimit-1]) + "…"
}
