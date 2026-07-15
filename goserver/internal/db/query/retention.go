package query

import (
	"context"
	"log"
	"time"

	"dealscanner/internal/models"

	"gorm.io/gorm"
)

type RetentionPolicy struct {
	ObservationRetention time.Duration
	ListingRetention     time.Duration
	TtlRetention         time.Duration
}

type RetentionPruneResult struct {
	ListingObservationsDeleted int64
	QueryListingStatesDeleted  int64
	ListingsDeleted            int64
	TtlItemsDeleted            int64
}

func (q *Query) StartRetentionPruner(ctx context.Context, interval time.Duration, policy RetentionPolicy) {
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	q.logRetentionPrune(ctx, policy)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.logRetentionPrune(ctx, policy)
		}
	}
}

func (q *Query) logRetentionPrune(ctx context.Context, policy RetentionPolicy) {
	result, err := q.PruneDatabase(ctx, time.Now().UTC(), policy)
	if err != nil {
		log.Printf("Database retention prune failed: %v", err)
		return
	}
	log.Printf(
		"Database retention prune complete: observations=%d states=%d listings=%d ttl=%d",
		result.ListingObservationsDeleted,
		result.QueryListingStatesDeleted,
		result.ListingsDeleted,
		result.TtlItemsDeleted,
	)
}

func (q *Query) PruneDatabase(ctx context.Context, now time.Time, policy RetentionPolicy) (RetentionPruneResult, error) {
	result := RetentionPruneResult{}

	err := q.UnderlyingDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if policy.ObservationRetention > 0 {
			observationCutoff := now.Add(-policy.ObservationRetention)
			db := tx.Where("observedAt < ?", observationCutoff).Delete(&models.ListingObservation{})
			if db.Error != nil {
				return db.Error
			}
			result.ListingObservationsDeleted = db.RowsAffected
		}

		if policy.ListingRetention > 0 {
			listingCutoff := now.Add(-policy.ListingRetention)
			db := tx.Where(
				"lastEvaluatedAt < ? AND (lastNotifiedAt IS NULL OR lastNotifiedAt < ?)",
				listingCutoff,
				listingCutoff,
			).Delete(&models.QueryListingState{})
			if db.Error != nil {
				return db.Error
			}
			result.QueryListingStatesDeleted = db.RowsAffected

			referencedStates := tx.Model(&models.QueryListingState{}).Select("listingId")
			referencedObservations := tx.Model(&models.ListingObservation{}).Select("listingId")
			db = tx.
				Where("lastSeenAt < ?", listingCutoff).
				Where("id NOT IN (?)", referencedStates).
				Where("id NOT IN (?)", referencedObservations).
				Delete(&models.Listing{})
			if db.Error != nil {
				return db.Error
			}
			result.ListingsDeleted = db.RowsAffected
		}

		if policy.TtlRetention > 0 {
			ttlCutoff := now.Add(-policy.TtlRetention)
			db := tx.Where("lastUpdated < ?", ttlCutoff).Delete(&models.TtlItem{})
			if db.Error != nil {
				return db.Error
			}
			result.TtlItemsDeleted = db.RowsAffected
		}

		return nil
	})

	return result, err
}
