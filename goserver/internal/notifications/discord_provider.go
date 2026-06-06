package notifications

import (
	"context"
	"fmt"

	"dealscanner/internal/config"
)

type BotSender interface {
	SendDeal(ctx context.Context, channelId, message string, imageUrl *string, queryType, queryId string) error
	SendError(channelId, message string) error
}

type DiscordProvider struct {
	cfg    *config.Config
	sender BotSender
}

func NewDiscordProvider(cfg *config.Config, sender BotSender) *DiscordProvider {
	return &DiscordProvider{
		cfg:    cfg,
		sender: sender,
	}
}

func (d *DiscordProvider) Name() string {
	return "discord"
}

func (d *DiscordProvider) IsEnabled() bool {
	return d.cfg.DiscordToken != "" && d.cfg.BotClientId != "" && d.cfg.DiscordGuildId != ""
}

func (d *DiscordProvider) Send(ctx context.Context, notification AppNotification) error {
	if notification.Kind == "error" {
		if d.cfg.ErrorChannelId == "" {
			return nil
		}
		msg := fmt.Sprintf("❌ **An error occurred in %s**:\n\n%s", notification.Source, notification.Message)
		if notification.Stack != "" {
			msg += fmt.Sprintf("\n\n```\n%s\n```", notification.Stack)
		}
		return d.sender.SendError(d.cfg.ErrorChannelId, msg)
	}

	channelId := d.getChannelId(notification.Source)
	roleId := d.getRoleId(notification.Source)
	if channelId == "" {
		return nil
	}

	prelude := "🔔 [DealScanner]" // simple preview prefix
	roleMention := ""
	if roleId != "" {
		roleMention = fmt.Sprintf("<@&%s> ", roleId)
	}

	message := fmt.Sprintf("%s%s %s", roleMention, prelude, notification.Title)
	if notification.Url != "" {
		message += fmt.Sprintf("\nLink: %s", notification.Url)
	}

	qType := ""
	qId := ""
	if notification.Query != nil {
		qType = notification.Query.Type
		qId = notification.Query.Id
	}

	return d.sender.SendDeal(ctx, channelId, message, notification.ImageUrl, qType, qId)
}

func (d *DiscordProvider) getChannelId(source string) string {
	switch source {
	case "cashConverters":
		return d.cfg.CashConvertersChannelId
	case "ebay":
		return d.cfg.EbayChannelId
	case "gumtree":
		return d.cfg.GumtreeChannelId
	case "salvos":
		return d.cfg.SalvosChannelId
	case "steamMarket":
		return d.cfg.SteamQueryChannelId
	case "csTrade", "lootFarm", "tradeIt":
		return d.cfg.CsChannelId
	default:
		return ""
	}
}

func (d *DiscordProvider) getRoleId(source string) string {
	switch source {
	case "cashConverters":
		return d.cfg.CashConvertersRoleId
	case "ebay":
		return d.cfg.EbayRoleId
	case "gumtree":
		return d.cfg.GumtreeRoleId
	case "salvos":
		return d.cfg.SalvosRoleId
	case "steamMarket":
		return d.cfg.SteamQueryRoleId
	case "csTrade", "lootFarm", "tradeIt":
		return d.cfg.CsRoleId
	default:
		return ""
	}
}
