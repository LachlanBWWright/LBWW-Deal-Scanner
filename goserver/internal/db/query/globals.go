package query

import (
	"context"
	"dealscanner/internal/config"
	"dealscanner/internal/models"
	"errors"

	"gorm.io/gorm"
)

func (q *Query) GetGlobals(ctx context.Context) (*models.Globals, error) {
	g, err := q.Globals.WithContext(ctx).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return g, nil
}

func (q *Query) UpdateGlobals(ctx context.Context, g *models.Globals) error {
	return q.Globals.WithContext(ctx).Save(g)
}

func (q *Query) SeedGlobals(ctx context.Context, cfg *config.Config) error {
	_, err := q.Globals.WithContext(ctx).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		g := models.Globals{
			ID:                      "1",
			BotClientId:             cfg.BotClientId,
			DiscordGuildId:          cfg.DiscordGuildId,
			DiscordToken:            cfg.DiscordToken,
			CashConverters:          cfg.CashConverters,
			CashConvertersChannelId: cfg.CashConvertersChannelId,
			CashConvertersRoleId:    cfg.CashConvertersRoleId,
			CommandPermissionRoleId: cfg.CommandPermissionRoleId,
			CsChannelId:             cfg.CsChannelId,
			CsItems:                 cfg.CsItems,
			CsMarketChannelId:       cfg.CsMarketChannelId,
			CsMarketRoleId:          cfg.CsMarketRoleId,
			CsRoleId:                cfg.CsRoleId,
			Ebay:                    cfg.Ebay,
			EbayChannelId:           cfg.EbayChannelId,
			EbayRoleId:              cfg.EbayRoleId,
			ErrorChannelId:          cfg.ErrorChannelId,
			Gumtree:                 cfg.Gumtree,
			GumtreeChannelId:        cfg.GumtreeChannelId,
			GumtreeRoleId:           cfg.GumtreeRoleId,
			Salvos:                  cfg.Salvos,
			SalvosChannelId:         cfg.SalvosChannelId,
			SalvosRoleId:            cfg.SalvosRoleId,
			SteamQuery:              cfg.SteamQuery,
			SteamQueryChannelId:     cfg.SteamQueryChannelId,
			SteamQueryRoleId:        cfg.SteamQueryRoleId,
			CsTradeDollarRatio:      1.0,
			LootFarmDollarRatio:     1.0,
			CsDealsDollarRatio:      1.0,
			TradeitGgDollarRatio:    1.0,
		}
		return q.Globals.WithContext(ctx).Create(&g)
	}
	return err
}
