package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	ScannerEbay           = 0
	ScannerGumtree        = 1
	ScannerSalvos         = 2
	ScannerCashConverters = 3
	ScannerSteamQuery     = 4
	ScannerCsTradeBot     = 5
)

type TtlItem struct {
	ItemId      string    `gorm:"primaryKey;column:itemId"`
	Scanner     int       `gorm:"primaryKey;column:scanner"`
	LastUpdated time.Time `gorm:"column:lastUpdated"`
}

func (TtlItem) TableName() string {
	return "TtlItem"
}

func (db *DB) CheckIfNew(ctx context.Context, itemId string, scanner int) (bool, error) {
	var item TtlItem
	err := db.WithContext(ctx).Where("itemId = ? AND scanner = ?", itemId, scanner).First(&item).Error
	if err == nil {
		if err := db.WithContext(ctx).Model(&item).Update("lastUpdated", time.Now().UTC()).Error; err != nil {
			return false, err
		}
		return false, nil
	}
	if err == gorm.ErrRecordNotFound {
		item = TtlItem{
			ItemId:      itemId,
			Scanner:     scanner,
			LastUpdated: time.Now().UTC(),
		}
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return false, err
		}
		return true, nil
	}
	return false, err
}

func (db *DB) CheckIfNewCsItem(ctx context.Context, itemName string, floatVal float64, csSite string) (bool, error) {
	itemId := fmt.Sprintf("%s_%f_%s", itemName, floatVal, csSite)
	return db.CheckIfNew(ctx, itemId, ScannerCsTradeBot)
}
