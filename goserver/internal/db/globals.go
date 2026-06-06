package db

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

func (db *DB) GetGlobals(ctx context.Context) (*Globals, error) {
	var g Globals
	if err := db.WithContext(ctx).First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (db *DB) UpdateGlobals(ctx context.Context, g *Globals) error {
	return db.WithContext(ctx).Save(g).Error
}
