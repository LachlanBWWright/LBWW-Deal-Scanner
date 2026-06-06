package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

func (db *DB) GetScannerRuntimeState(ctx context.Context) (*ScannerRuntimeState, error) {
	var s ScannerRuntimeState
	err := db.WithContext(ctx).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s = ScannerRuntimeState{
			ID:        "1",
			UpdatedAt: time.Now().UTC(),
		}
		if err := db.WithContext(ctx).Create(&s).Error; err != nil {
			return nil, err
		}
		return &s, nil
	}
	return &s, err
}

func (db *DB) UpdateScannerRuntimeState(ctx context.Context, s *ScannerRuntimeState) error {
	s.UpdatedAt = time.Now().UTC()
	return db.WithContext(ctx).Save(s).Error
}
