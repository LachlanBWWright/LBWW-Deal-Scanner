package notifications

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"dealscanner/internal/config"
)

type BotSender interface {
	SendDeal(ctx context.Context, channelId, channelMessage, dmMessage string, imageUrl *string, queryType, queryId string) error
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

var standardPreludes = []string{
	"Listen up, Jack,",
	"My fellow Americans,",
	"Folks,",
	"Here's the deal,",
}

var rareNotificationPreludes = []string{
	"This is a big fu- ...uh... flippable deal,",
}

func getNotificationPrelude() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Float64() < 0.01 {
		return rareNotificationPreludes[r.Intn(len(rareNotificationPreludes))]
	}
	return standardPreludes[r.Intn(len(standardPreludes))]
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

	prelude := getNotificationPrelude()
	roleMention := ""
	if roleId != "" {
		roleMention = fmt.Sprintf("<@&%s> ", roleId)
	}

	dmMessage := fmt.Sprintf("%s %s", prelude, notification.Title)
	channelMessage := roleMention + dmMessage

	qType := ""
	qId := ""
	if notification.Query != nil {
		qType = notification.Query.Type
		qId = notification.Query.Id
	}

	return d.sender.SendDeal(ctx, channelId, channelMessage, dmMessage, notification.ImageUrl, qType, qId)
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
