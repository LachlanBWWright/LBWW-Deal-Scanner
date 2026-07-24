package discord

import (
	"fmt"
	"strings"

	"dealscanner/internal/config"
)

func BuildStartupMessage(cfg *config.Config) string {
	channels := []namedDiscordID{
		{name: "Logs", id: cfg.ErrorChannelId},
		{name: "Cash Converters", id: cfg.CashConvertersChannelId},
		{name: "eBay", id: cfg.EbayChannelId},
		{name: "Gumtree", id: cfg.GumtreeChannelId},
		{name: "Salvos", id: cfg.SalvosChannelId},
		{name: "Steam Market", id: cfg.SteamQueryChannelId},
		{name: "CS", id: cfg.CsChannelId},
		{name: "CS Market", id: cfg.CsMarketChannelId},
	}
	roles := []namedDiscordID{
		{name: "Command permission", id: cfg.CommandPermissionRoleId},
		{name: "Cash Converters", id: cfg.CashConvertersRoleId},
		{name: "eBay", id: cfg.EbayRoleId},
		{name: "Gumtree", id: cfg.GumtreeRoleId},
		{name: "Salvos", id: cfg.SalvosRoleId},
		{name: "Steam Market", id: cfg.SteamQueryRoleId},
		{name: "CS", id: cfg.CsRoleId},
		{name: "CS Market", id: cfg.CsMarketRoleId},
	}
	return "DealScanner started\n\n**Configured channels**\n" +
		formatDiscordIDs(channels, "channel") +
		"\n\n**Configured roles**\n" +
		formatDiscordIDs(roles, "role")
}

type namedDiscordID struct {
	name string
	id   string
}

func formatDiscordIDs(values []namedDiscordID, kind string) string {
	lines := make([]string, 0, len(values))
	for _, value := range values {
		if value.id == "" {
			continue
		}
		mention := fmt.Sprintf("<@&%s>", value.id)
		if kind == "channel" {
			mention = fmt.Sprintf("<#%s>", value.id)
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", value.name, mention))
	}
	if len(lines) == 0 {
		return "- None"
	}
	return strings.Join(lines, "\n")
}
