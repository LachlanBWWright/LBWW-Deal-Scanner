package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(connStr string) (*gorm.DB, error) {
	cleanPath := connStr
	if strings.HasPrefix(cleanPath, "file:") {
		cleanPath = strings.TrimPrefix(cleanPath, "file:")
		if idx := strings.Index(cleanPath, "?"); idx != -1 {
			cleanPath = cleanPath[:idx]
		}
	}

	// Create directory if not exists
	dir := filepath.Dir(cleanPath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	}

	db, err := gorm.Open(sqlite.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migrate all database models
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
		return nil, fmt.Errorf("failed to run GORM auto-migrations: %w", err)
	}

	return db, nil
}
