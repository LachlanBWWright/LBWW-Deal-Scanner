package query

import (
	"context"
	"dealscanner/internal/config"
	"dealscanner/internal/models"
	"errors"

	"gorm.io/gorm"
)

func (q *Query) GetGlobals(ctx context.Context) (*models.Globals, error) {
	if cached, ok := defaultGlobalsCache.get(); ok {
		return cached, nil
	}

	g, err := q.Globals.WithContext(ctx).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	defaultGlobalsCache.set(g)
	return g, nil
}

func (q *Query) UpdateGlobals(ctx context.Context, g *models.Globals) error {
	if err := q.Globals.WithContext(ctx).Save(g); err != nil {
		return err
	}
	defaultGlobalsCache.set(g)
	return nil
}

func (q *Query) SeedGlobals(ctx context.Context, cfg *config.Config) error {
	g, err := q.Globals.WithContext(ctx).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		g = &models.Globals{
			ID:                   "1",
			CsTradeDollarRatio:   1.0,
			LootFarmDollarRatio:  1.0,
			CsDealsDollarRatio:   1.0,
			TradeitGgDollarRatio: 1.0,
		}
	} else if err != nil {
		return err
	}

	// The environment is the deployment source of truth for scanner switches,
	// Discord credentials, channels, and roles. Keep an existing migrated row
	// synchronized instead of retaining stale defaults indefinitely.
	g.BotClientId = cfg.BotClientId
	g.DiscordGuildId = cfg.DiscordGuildId
	g.DiscordToken = cfg.DiscordToken
	g.CashConverters = cfg.CashConverters
	g.CashConvertersChannelId = cfg.CashConvertersChannelId
	g.CashConvertersRoleId = cfg.CashConvertersRoleId
	g.CommandPermissionRoleId = cfg.CommandPermissionRoleId
	g.CsChannelId = cfg.CsChannelId
	g.CsItems = cfg.CsItems
	g.CsMarketChannelId = cfg.CsMarketChannelId
	g.CsMarketRoleId = cfg.CsMarketRoleId
	g.CsRoleId = cfg.CsRoleId
	g.Ebay = cfg.Ebay
	g.EbayChannelId = cfg.EbayChannelId
	g.EbayRoleId = cfg.EbayRoleId
	g.ErrorChannelId = cfg.ErrorChannelId
	g.Gumtree = cfg.Gumtree
	g.GumtreeChannelId = cfg.GumtreeChannelId
	g.GumtreeRoleId = cfg.GumtreeRoleId
	g.Salvos = cfg.Salvos
	g.SalvosChannelId = cfg.SalvosChannelId
	g.SalvosRoleId = cfg.SalvosRoleId
	g.SteamQuery = cfg.SteamQuery
	g.SteamQueryChannelId = cfg.SteamQueryChannelId
	g.SteamQueryRoleId = cfg.SteamQueryRoleId

	return q.UpdateGlobals(ctx, g)
}
