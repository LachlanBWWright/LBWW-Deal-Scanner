package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

func (db *DB) GetAction(ctx context.Context, key string) (*ActionRegistry, error) {
	var a ActionRegistry
	if err := db.WithContext(ctx).Where("id = ?", key).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (db *DB) SetAction(ctx context.Context, key string, a ActionRegistry) error {
	a.ID = key
	a.CreatedAt = time.Now().UTC()
	return db.WithContext(ctx).Save(&a).Error
}

func (db *DB) DeleteAction(ctx context.Context, key string) (bool, error) {
	res := db.WithContext(ctx).Where("id = ?", key).Delete(&ActionRegistry{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (db *DB) CleanupExpiredActions(ctx context.Context, maxAgeMs int64) error {
	cutoff := time.Now().UnixMilli() - maxAgeMs
	return db.WithContext(ctx).Where("timestamp < ?", cutoff).Delete(&ActionRegistry{}).Error
}
