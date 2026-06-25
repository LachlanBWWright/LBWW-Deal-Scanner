package query

import (
	"context"
	"testing"
	"time"

	"dealscanner/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeleteSavedQueryRemovesDependentRecords(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file:delete-query-test?mode=memory&cache=shared&_foreign_keys=on"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	err = gdb.AutoMigrate(
		&models.SearchQuery{},
		&models.Ebay{},
		&models.UserQuery{},
		&models.Listing{},
		&models.QueryListingState{},
	)
	if err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	dbClient := Use(gdb)
	ctx := context.Background()
	queryID := "query-to-delete"
	listingID := "listing-for-query"

	if err = dbClient.SearchQuery.WithContext(ctx).Create(&models.SearchQuery{
		ID:        queryID,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create parent query: %v", err)
	}
	if err = dbClient.Ebay.WithContext(ctx).Create(&models.Ebay{
		Url:     "https://example.test/query",
		QueryId: queryID,
	}); err != nil {
		t.Fatalf("create ebay query: %v", err)
	}
	if err = dbClient.UserQuery.WithContext(ctx).Create(&models.UserQuery{
		ID:        "user-query",
		UserId:    "user",
		QueryId:   queryID,
		QueryType: "ebay",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create user query: %v", err)
	}
	if err = dbClient.Listing.WithContext(ctx).Create(&models.Listing{
		ID:           listingID,
		Source:       "ebay",
		CanonicalUrl: "https://example.test/listing",
		Title:        "Listing",
		FirstSeenAt:  time.Now().UTC(),
		LastSeenAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create listing: %v", err)
	}
	if err = dbClient.QueryListingState.WithContext(ctx).Create(&models.QueryListingState{
		QueryId:         queryID,
		ListingId:       listingID,
		Source:          "ebay",
		Status:          ListingStateStatusMatched,
		LastEvaluatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create query listing state: %v", err)
	}

	if err = dbClient.DeleteSavedQuery(ctx, "ebay", "https://example.test/query"); err != nil {
		t.Fatalf("delete saved query: %v", err)
	}

	for _, table := range []string{"Query", "Ebay", "UserQuery", "QueryListingState"} {
		var count int64
		if err = gdb.Table(table).Count(&count).Error; err != nil {
			t.Fatalf("count %s records: %v", table, err)
		}
		if count != 0 {
			t.Errorf("expected %s records to be deleted, found %d", table, count)
		}
	}
}
