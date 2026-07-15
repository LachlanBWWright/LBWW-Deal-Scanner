package discord

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"dealscanner/internal/models"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

func (b *Bot) handleButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	customId := i.MessageComponentData().CustomID

	action, err := b.dbClient.GetAction(ctx, customId)
	if err != nil || action == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Action expired or not found.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if action.Type == "view_page" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
	}

	var responseContent string
	var responseComponents []discordgo.MessageComponent

	switch action.Type {
	case "delete":
		confirmKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		cancelKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		now := time.Now().UnixMilli()

		if err := b.dbClient.SetAction(ctx, confirmKey, models.ActionRegistry{
			ID:         confirmKey,
			Type:       "confirm_delete",
			QueryType:  action.QueryType,
			QueryId:    action.QueryId,
			UserId:     &i.Member.User.ID,
			Timestamp:  now,
			RelatedKey: &customId,
		}); err != nil {
			log.Printf("Failed to register Discord confirm delete action: %v", err)
		}

		if err := b.dbClient.SetAction(ctx, cancelKey, models.ActionRegistry{
			ID:         cancelKey,
			Type:       "cancel_delete",
			QueryType:  action.QueryType,
			QueryId:    action.QueryId,
			UserId:     &i.Member.User.ID,
			Timestamp:  now,
			RelatedKey: &customId,
		}); err != nil {
			log.Printf("Failed to register Discord cancel delete action: %v", err)
		}

		responseContent = fmt.Sprintf("Are you sure you want to delete this %s query?\n%s\n\n**This action cannot be undone.**", *action.QueryType, formatQueryReference(*action.QueryId))
		responseComponents = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Yes, Delete Query",
						Style:    discordgo.DangerButton,
						CustomID: confirmKey,
					},
					discordgo.Button{
						Label:    "Cancel",
						Style:    discordgo.SecondaryButton,
						CustomID: cancelKey,
					},
				},
			},
		}

	case "confirm_delete":
		if action.QueryType != nil && action.QueryId != nil {
			err = b.dbClient.DeleteSavedQuery(ctx, *action.QueryType, *action.QueryId)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error deleting query: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Successfully deleted query of type `%s` (%s)", *action.QueryType, formatQueryReference(*action.QueryId))
				b.dbClient.DeleteAction(ctx, customId)
				if action.RelatedKey != nil {
					b.dbClient.DeleteAction(ctx, *action.RelatedKey)
				}
			}
		}

	case "cancel_delete":
		responseContent = "Deletion cancelled."
		b.dbClient.DeleteAction(ctx, customId)
		if action.RelatedKey != nil {
			b.dbClient.DeleteAction(ctx, *action.RelatedKey)
		}

	case "subscribe_dm":
		var memberID string
		if i.Member != nil {
			memberID = i.Member.User.ID
		} else if i.User != nil {
			memberID = i.User.ID
		}

		parentQueryId, err := b.getQueryIdByTypeAndKey(ctx, *action.QueryType, *action.QueryId)
		if err != nil || parentQueryId == "" {
			responseContent = "❌ Query not found."
		} else {
			count, err := b.dbClient.UserQuery.WithContext(ctx).Where(b.dbClient.UserQuery.UserId.Eq(memberID), b.dbClient.UserQuery.QueryId.Eq(parentQueryId)).Count()
			if err == nil && count > 0 {
				responseContent = "✅ You are already subscribed to DMs for this query."
			} else {
				uq := models.UserQuery{
					ID:        uuid.New().String(),
					UserId:    memberID,
					QueryId:   parentQueryId,
					QueryType: *action.QueryType,
					CreatedAt: time.Now().UTC(),
				}
				if err := b.dbClient.UserQuery.WithContext(ctx).Create(&uq); err != nil {
					responseContent = fmt.Sprintf("❌ Failed to subscribe: %v", err)
				} else {
					responseContent = fmt.Sprintf("✅ You have been subscribed to DM notifications for this %s query.", *action.QueryType)
					b.dbClient.DeleteAction(ctx, customId)
				}
			}
		}

	case "unsubscribe_dm":
		var memberID string
		if i.Member != nil {
			memberID = i.Member.User.ID
		} else if i.User != nil {
			memberID = i.User.ID
		}

		parentQueryId, err := b.getQueryIdByTypeAndKey(ctx, *action.QueryType, *action.QueryId)
		if err != nil || parentQueryId == "" {
			responseContent = "❌ Query not found."
		} else {
			res, err := b.dbClient.UserQuery.WithContext(ctx).Where(b.dbClient.UserQuery.UserId.Eq(memberID), b.dbClient.UserQuery.QueryId.Eq(parentQueryId)).Delete()
			if err != nil {
				responseContent = fmt.Sprintf("❌ Failed to unsubscribe: %v", err)
			} else if res.RowsAffected == 0 {
				responseContent = "❌ You are not subscribed to this query."
			} else {
				responseContent = "✅ You have been unsubscribed from DM notifications for this query."
				b.dbClient.DeleteAction(ctx, customId)
			}
		}

	case "view_query_details":
		if action.QueryType == nil || action.QueryId == nil {
			responseContent = "❌ Query details are unavailable."
			break
		}
		responseContent = b.formatQueryDetails(ctx, *action.QueryType, *action.QueryId)

	case "view_page":
		if action.QueryType != nil && action.QueryId != nil {
			pageVal, err := strconv.Atoi(*action.QueryId)
			if err == nil {
				title := ""
				switch *action.QueryType {
				case "ebay":
					title = "Saved eBay Queries"
				case "cashConverters":
					title = "Saved Cash Converters Queries"
				case "gumtree":
					title = "Saved Gumtree Queries"
				case "salvos":
					title = "Saved Salvos Queries"
				case "csMarket":
					title = "Saved CS Market Queries"
				case "csTradeBot":
					title = "Saved CS Trade Bot Multisearches"
				case "steamMarket":
					title = "Saved Steam SCM Queries"
				}
				b.sendPaginatedQueries(ctx, s, i, *action.QueryType, title, pageVal)
				b.dbClient.DeleteAction(ctx, customId)
				return
			}
		}
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &responseContent,
		Components: &responseComponents,
	})
}

// SendDeal publishes a deal notification to the configured channel and subscribed users
