package main

import (
	"fmt"
	"log"
	"strings"

	"dealscanner/internal/config"
	"dealscanner/internal/models"
	"gorm.io/gorm"
)

type foreignKey struct {
	Table    string `gorm:"column:table"`
	From     string `gorm:"column:from"`
	OnDelete string `gorm:"column:on_delete"`
}

type expectedForeignKey struct {
	table      string
	column     string
	references string
}

func hasCascadeConstraint(
	db *gorm.DB,
	table string,
	column string,
	references string,
) (bool, error) {
	var foreignKeys []foreignKey
	query := fmt.Sprintf("PRAGMA foreign_key_list(%q)", table)
	if err := db.Raw(query).Scan(&foreignKeys).Error; err != nil {
		return false, fmt.Errorf("inspect foreign keys for %s: %w", table, err)
	}

	for _, foreignKey := range foreignKeys {
		if foreignKey.From == column &&
			foreignKey.Table == references &&
			strings.EqualFold(foreignKey.OnDelete, "CASCADE") {
			return true, nil
		}
	}

	return false, nil
}

func repairCascadeConstraint(
	db *gorm.DB,
	table string,
	column string,
	references string,
	drop func() error,
	create func() error,
) error {
	hasCascade, err := hasCascadeConstraint(db, table, column, references)
	if err != nil {
		return err
	}
	if hasCascade {
		return nil
	}

	if err := drop(); err != nil {
		return fmt.Errorf("drop %s.%s foreign key: %w", table, column, err)
	}
	if err := create(); err != nil {
		return fmt.Errorf("create %s.%s cascade foreign key: %w", table, column, err)
	}

	return nil
}

func migrateCascadeConstraints(db *gorm.DB) error {
	migrator := db.Migrator()
	migrations := []func() error{
		func() error {
			return repairCascadeConstraint(db, "UserQuery", "queryId", "Query",
				func() error { return migrator.DropConstraint(&models.UserQuery{}, "Query") },
				func() error { return migrator.CreateConstraint(&models.UserQuery{}, "Query") },
			)
		},
		func() error {
			return repairCascadeConstraint(db, "CashConvertersFilter", "queryId", "Query",
				func() error { return migrator.DropConstraint(&models.CashConvertersFilter{}, "Query") },
				func() error { return migrator.CreateConstraint(&models.CashConvertersFilter{}, "Query") },
			)
		},
		func() error {
			return repairCascadeConstraint(db, "Ebay", "queryId", "Query",
				func() error { return migrator.DropConstraint(&models.Ebay{}, "Query") },
				func() error { return migrator.CreateConstraint(&models.Ebay{}, "Query") },
			)
		},
		func() error {
			return repairCascadeConstraint(db, "Gumtree", "queryId", "Query",
				func() error { return migrator.DropConstraint(&models.Gumtree{}, "Query") },
				func() error { return migrator.CreateConstraint(&models.Gumtree{}, "Query") },
			)
		},
		func() error {
			return repairCascadeConstraint(db, "Salvos", "queryId", "Query",
				func() error { return migrator.DropConstraint(&models.Salvos{}, "Query") },
				func() error { return migrator.CreateConstraint(&models.Salvos{}, "Query") },
			)
		},
		func() error {
			return repairCascadeConstraint(db, "CsMarket", "queryId", "Query",
				func() error { return migrator.DropConstraint(&models.CsMarket{}, "Query") },
				func() error { return migrator.CreateConstraint(&models.CsMarket{}, "Query") },
			)
		},
		func() error {
			return repairCascadeConstraint(db, "SteamMarket", "queryId", "Query",
				func() error { return migrator.DropConstraint(&models.SteamMarket{}, "Query") },
				func() error { return migrator.CreateConstraint(&models.SteamMarket{}, "Query") },
			)
		},
		func() error {
			return repairCascadeConstraint(db, "CsTradeBot", "queryId", "Query",
				func() error { return migrator.DropConstraint(&models.CsTradeBot{}, "Query") },
				func() error { return migrator.CreateConstraint(&models.CsTradeBot{}, "Query") },
			)
		},
		func() error {
			return repairCascadeConstraint(db, "QueryListingState", "queryId", "Query",
				func() error { return migrator.DropConstraint(&models.QueryListingState{}, "Query") },
				func() error { return migrator.CreateConstraint(&models.QueryListingState{}, "Query") },
			)
		},
		func() error {
			return repairCascadeConstraint(db, "QueryListingState", "listingId", "Listing",
				func() error { return migrator.DropConstraint(&models.QueryListingState{}, "Listing") },
				func() error { return migrator.CreateConstraint(&models.QueryListingState{}, "Listing") },
			)
		},
	}

	for _, migrate := range migrations {
		if err := migrate(); err != nil {
			return err
		}
	}

	return nil
}

func verifyCascadeConstraints(db *gorm.DB) error {
	expected := []expectedForeignKey{
		{table: "UserQuery", column: "queryId", references: "Query"},
		{table: "CashConvertersFilter", column: "queryId", references: "Query"},
		{table: "Ebay", column: "queryId", references: "Query"},
		{table: "Gumtree", column: "queryId", references: "Query"},
		{table: "Salvos", column: "queryId", references: "Query"},
		{table: "CsMarket", column: "queryId", references: "Query"},
		{table: "SteamMarket", column: "queryId", references: "Query"},
		{table: "CsTradeBot", column: "queryId", references: "Query"},
		{table: "QueryListingState", column: "queryId", references: "Query"},
		{table: "QueryListingState", column: "listingId", references: "Listing"},
	}

	foreignKeysByTable := make(map[string][]foreignKey)
	for _, constraint := range expected {
		if _, exists := foreignKeysByTable[constraint.table]; exists {
			continue
		}

		var foreignKeys []foreignKey
		query := fmt.Sprintf("PRAGMA foreign_key_list(%q)", constraint.table)
		if err := db.Raw(query).Scan(&foreignKeys).Error; err != nil {
			return fmt.Errorf("inspect foreign keys for %s: %w", constraint.table, err)
		}
		foreignKeysByTable[constraint.table] = foreignKeys
	}

	for _, constraint := range expected {
		found := false
		for _, foreignKey := range foreignKeysByTable[constraint.table] {
			if foreignKey.From == constraint.column &&
				foreignKey.Table == constraint.references &&
				strings.EqualFold(foreignKey.OnDelete, "CASCADE") {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf(
				"%s.%s must reference %s with ON DELETE CASCADE",
				constraint.table,
				constraint.column,
				constraint.references,
			)
		}
	}

	return nil
}

func main() {
	cfg := config.LoadConfig()

	db, err := models.Open(cfg.TursoDatabaseUrl, cfg.TursoAuthToken)
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to access SQL database handle: %v", err)
	}
	defer sqlDB.Close()

	if err := migrateCascadeConstraints(db); err != nil {
		log.Fatalf("Database cascade migration failed: %v", err)
	}

	if err := verifyCascadeConstraints(db); err != nil {
		log.Fatalf("Database cascade verification failed: %v", err)
	}

	log.Println("Turso database schema migration completed.")
}
