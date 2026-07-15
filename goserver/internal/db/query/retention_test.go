package query

import (
	"context"
	"testing"
	"time"

	"dealscanner/internal/models"
	"gorm.io/gorm"
)

func TestPruneDatabaseRemovesStaleRowsAndKeepsCurrentReferences(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	db := dbClient.UnderlyingDB().WithContext(ctx)
	if err := db.AutoMigrate(&models.TtlItem{}, &models.ActionRegistry{}); err != nil {
		t.Fatalf("migrate retention models: %v", err)
	}

	now := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	queryObj := models.SearchQuery{ID: "retention-query", CreatedAt: now.Add(-90 * 24 * time.Hour)}
	if err := db.Create(&queryObj).Error; err != nil {
		t.Fatalf("create query: %v", err)
	}

	oldUnreferenced := models.Listing{
		ID:           "old-unreferenced",
		Source:       "ebay",
		CanonicalUrl: "https://example.test/old-unreferenced",
		Title:        "Old unreferenced",
		FirstSeenAt:  now.Add(-30 * 24 * time.Hour),
		LastSeenAt:   now.Add(-10 * 24 * time.Hour),
	}
	currentObservationListing := models.Listing{
		ID:           "current-observation",
		Source:       "ebay",
		CanonicalUrl: "https://example.test/current-observation",
		Title:        "Current observation",
		FirstSeenAt:  now.Add(-30 * 24 * time.Hour),
		LastSeenAt:   now.Add(-10 * 24 * time.Hour),
	}
	currentStateListing := models.Listing{
		ID:           "current-state",
		Source:       "ebay",
		CanonicalUrl: "https://example.test/current-state",
		Title:        "Current state",
		FirstSeenAt:  now.Add(-30 * 24 * time.Hour),
		LastSeenAt:   now.Add(-10 * 24 * time.Hour),
	}
	if err := db.Create(&[]models.Listing{oldUnreferenced, currentObservationListing, currentStateListing}).Error; err != nil {
		t.Fatalf("create listings: %v", err)
	}

	observations := []models.ListingObservation{
		{
			ID:           "old-observation",
			ListingId:    oldUnreferenced.ID,
			Source:       "ebay",
			ObservedAt:   now.Add(-10 * 24 * time.Hour),
			Title:        "Old observation",
			Availability: stringPtr("available"),
		},
		{
			ID:           "current-observation",
			ListingId:    currentObservationListing.ID,
			Source:       "ebay",
			ObservedAt:   now.Add(-2 * 24 * time.Hour),
			Title:        "Current observation",
			Availability: stringPtr("available"),
		},
	}
	if err := db.Create(&observations).Error; err != nil {
		t.Fatalf("create observations: %v", err)
	}

	oldNotifiedAt := now.Add(-10 * 24 * time.Hour)
	currentNotifiedAt := now.Add(-2 * 24 * time.Hour)
	states := []models.QueryListingState{
		{
			QueryId:         queryObj.ID,
			ListingId:       oldUnreferenced.ID,
			Source:          "ebay",
			Status:          models.QueryListingStateStatusRejected,
			LastEvaluatedAt: now.Add(-10 * 24 * time.Hour),
			LastNotifiedAt:  &oldNotifiedAt,
		},
		{
			QueryId:         queryObj.ID,
			ListingId:       currentStateListing.ID,
			Source:          "ebay",
			Status:          models.QueryListingStateStatusMatched,
			LastEvaluatedAt: now.Add(-10 * 24 * time.Hour),
			LastNotifiedAt:  &currentNotifiedAt,
		},
	}
	if err := db.Create(&states).Error; err != nil {
		t.Fatalf("create states: %v", err)
	}

	ttlItems := []models.TtlItem{
		{ItemId: "old-ttl", Scanner: ScannerEbay, LastUpdated: now.Add(-10 * 24 * time.Hour)},
		{ItemId: "current-ttl", Scanner: ScannerEbay, LastUpdated: now.Add(-2 * 24 * time.Hour)},
	}
	if err := db.Create(&ttlItems).Error; err != nil {
		t.Fatalf("create ttl items: %v", err)
	}

	actions := []models.ActionRegistry{
		{ID: "old-action", Type: "test", CreatedAt: now.Add(-10 * 24 * time.Hour), Timestamp: now.Add(-10 * 24 * time.Hour).UnixMilli()},
		{ID: "current-action", Type: "test", CreatedAt: now.Add(-2 * 24 * time.Hour), Timestamp: now.Add(-2 * 24 * time.Hour).UnixMilli()},
	}
	if err := db.Create(&actions).Error; err != nil {
		t.Fatalf("create actions: %v", err)
	}

	result, err := dbClient.PruneDatabase(ctx, now, RetentionPolicy{
		ObservationRetention: 7 * 24 * time.Hour,
		ListingRetention:     7 * 24 * time.Hour,
		TtlRetention:         7 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("prune database: %v", err)
	}
	if result.ListingObservationsDeleted != 1 {
		t.Fatalf("observations deleted = %d; expected 1", result.ListingObservationsDeleted)
	}
	if result.QueryListingStatesDeleted != 1 {
		t.Fatalf("states deleted = %d; expected 1", result.QueryListingStatesDeleted)
	}
	if result.ListingsDeleted != 1 {
		t.Fatalf("listings deleted = %d; expected 1", result.ListingsDeleted)
	}
	if result.TtlItemsDeleted != 1 {
		t.Fatalf("ttl deleted = %d; expected 1", result.TtlItemsDeleted)
	}
	assertCount(t, db, &models.Listing{}, "id = ?", oldUnreferenced.ID, 0)
	assertCount(t, db, &models.Listing{}, "id = ?", currentObservationListing.ID, 1)
	assertCount(t, db, &models.Listing{}, "id = ?", currentStateListing.ID, 1)
	assertCount(t, db, &models.ListingObservation{}, "id = ?", "old-observation", 0)
	assertCount(t, db, &models.ListingObservation{}, "id = ?", "current-observation", 1)
	assertCount(t, db, &models.QueryListingState{}, "listingId = ?", oldUnreferenced.ID, 0)
	assertCount(t, db, &models.QueryListingState{}, "listingId = ?", currentStateListing.ID, 1)
	assertCount(t, db, &models.TtlItem{}, "itemId = ?", "old-ttl", 0)
	assertCount(t, db, &models.TtlItem{}, "itemId = ?", "current-ttl", 1)
	assertCount(t, db, &models.ActionRegistry{}, "id = ?", "old-action", 1)
	assertCount(t, db, &models.ActionRegistry{}, "id = ?", "current-action", 1)
}

func assertCount(t *testing.T, db *gorm.DB, model interface{}, where string, value string, expected int64) {
	t.Helper()
	var count int64
	if err := db.Model(model).Where(where, value).Count(&count).Error; err != nil {
		t.Fatalf("count %T where %s = %q: %v", model, where, value, err)
	}
	if count != expected {
		t.Fatalf("count %T where %s = %q is %d; expected %d", model, where, value, count, expected)
	}
}
