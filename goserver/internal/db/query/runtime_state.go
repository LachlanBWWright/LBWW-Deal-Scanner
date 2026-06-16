package query

import (
	"context"
	"errors"
	"time"

	"dealscanner/internal/models"
	"gorm.io/gorm"
)

func (q *Query) GetScannerRuntimeState(ctx context.Context) (*models.ScannerRuntimeState, error) {
	s, err := q.ScannerRuntimeState.WithContext(ctx).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newState := models.ScannerRuntimeState{
			ID:        "1",
			UpdatedAt: time.Now().UTC(),
		}
		if err := q.ScannerRuntimeState.WithContext(ctx).Create(&newState); err != nil {
			return nil, err
		}
		return &newState, nil
	}
	return s, err
}

func (q *Query) UpdateScannerRuntimeState(ctx context.Context, s *models.ScannerRuntimeState) error {
	s.UpdatedAt = time.Now().UTC()
	return q.ScannerRuntimeState.WithContext(ctx).Save(s)
}
