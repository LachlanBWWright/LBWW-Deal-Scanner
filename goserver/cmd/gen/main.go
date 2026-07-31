package main

import (
	"dealscanner/internal/models"
	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "internal/db/query",
		Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.ApplyBasic(
		&models.SearchQuery{},
		&models.UserQuery{},
		&models.CashConverters{},
		&models.CashConvertersFilter{},
		&models.Ebay{},
		&models.Gumtree{},
		&models.Salvos{},
		&models.CsMarket{},
		&models.SteamMarket{},
		&models.CsTradeBot{},
		&models.Globals{},
		&models.Listing{},
		&models.ListingObservation{},
		&models.QueryListingState{},
		&models.ScannerRuntimeState{},
		&models.ActionRegistry{},
		&models.TtlItem{},
		&models.CashConvertersScanState{},
		&models.CashConvertersListingMeta{},
		&models.CashConvertersDetailJob{},
		&models.CashConvertersSearchDoc{},
		&models.CashConvertersDeletedListing{},
	)

	g.Execute()
}
