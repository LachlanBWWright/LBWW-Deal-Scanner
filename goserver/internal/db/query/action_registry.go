package query

import (
	"context"
	"dealscanner/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

func (q *Query) GetAction(ctx context.Context, key string) (*models.ActionRegistry, error) {
	ar := q.ActionRegistry
	a, err := ar.WithContext(ctx).Where(ar.ID.Eq(key)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return a, nil
}

func (q *Query) SetAction(ctx context.Context, key string, a models.ActionRegistry) error {
	a.ID = key
	a.CreatedAt = time.Now().UTC()
	return q.ActionRegistry.WithContext(ctx).Save(&a)
}

func (q *Query) DeleteAction(ctx context.Context, key string) (bool, error) {
	ar := q.ActionRegistry
	res, err := ar.WithContext(ctx).Where(ar.ID.Eq(key)).Delete()
	if err != nil {
		return false, err
	}
	return res.RowsAffected > 0, nil
}

func (q *Query) CleanupExpiredActions(ctx context.Context, maxAgeMs int64) error {
	cutoff := time.Now().UnixMilli() - maxAgeMs
	ar := q.ActionRegistry
	_, err := ar.WithContext(ctx).Where(ar.Timestamp.Lt(cutoff)).Delete()
	return err
}
