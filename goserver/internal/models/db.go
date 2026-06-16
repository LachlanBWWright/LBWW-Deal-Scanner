package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(connStr string, tursoAuthToken string) (*gorm.DB, error) {
	if connStr == "" {
		return nil, fmt.Errorf("TURSO_DATABASE_URL is required")
	}
	if tursoAuthToken == "" {
		return nil, fmt.Errorf("TURSO_AUTH_TOKEN is required")
	}
	if !strings.HasPrefix(connStr, "libsql://") && !strings.HasPrefix(connStr, "https://") {
		return nil, fmt.Errorf("TURSO_DATABASE_URL must start with libsql:// or https://")
	}

	connector, err := libsql.NewConnector(connStr, libsql.WithAuthToken(tursoAuthToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create libsql connector: %w", err)
	}

	sqlDB := sql.OpenDB(connector)
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to open libsql database: %w", err)
	}

	err = db.AutoMigrate(
		&SearchQuery{},
		&UserQuery{},
		&CashConverters{},
		&Ebay{},
		&Gumtree{},
		&Salvos{},
		&CsMarket{},
		&SteamMarket{},
		&CsTradeBot{},
		&Globals{},
		&Listing{},
		&ListingObservation{},
		&QueryListingState{},
		&ScannerRuntimeState{},
		&ActionRegistry{},
		&TtlItem{},
	)
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run GORM auto-migrations: %w", err)
	}

	return db, nil
}
