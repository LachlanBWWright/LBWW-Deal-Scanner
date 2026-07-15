package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dealscanner/internal/models"
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

const ttlRefreshWriteInterval = time.Hour

func (q *Query) CheckIfNew(ctx context.Context, itemId string, scanner int) (bool, error) {
	item, err := q.TtlItem.WithContext(ctx).Where(q.TtlItem.ItemId.Eq(itemId), q.TtlItem.Scanner.Eq(scanner)).First()
	if err == nil {
		now := time.Now().UTC()
		if now.Sub(item.LastUpdated) >= ttlRefreshWriteInterval {
			if _, err := q.TtlItem.WithContext(ctx).Where(q.TtlItem.ItemId.Eq(itemId), q.TtlItem.Scanner.Eq(scanner)).Update(q.TtlItem.LastUpdated, now); err != nil {
				return false, err
			}
		}
		return false, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item := models.TtlItem{
			ItemId:      itemId,
			Scanner:     scanner,
			LastUpdated: time.Now().UTC(),
		}
		if err := q.TtlItem.WithContext(ctx).Create(&item); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, err
}

func (q *Query) CheckIfNewCsItem(ctx context.Context, itemName string, floatVal float64, csSite string) (bool, error) {
	itemId := fmt.Sprintf("%s_%f_%s", itemName, floatVal, csSite)
	return q.CheckIfNew(ctx, itemId, ScannerCsTradeBot)
}
