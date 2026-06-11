package query

import (
	"context"
	"testing"
	"time"

	"dealscanner/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupQueryCacheTestDB(t *testing.T) (*Query, func()) {
	t.Helper()

	defaultSavedQueryCache.invalidate()

	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = gdb.AutoMigrate(&models.SearchQuery{}, &models.Ebay{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbClient := Use(gdb)
	cleanup := func() {
		defaultSavedQueryCache.invalidate()
		sqlDB, _ := gdb.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return dbClient, cleanup
}

func TestListSavedQueriesUsesCacheAndReturnsDefensiveCopies(t *testing.T) {
	dbClient, cleanup := setupQueryCacheTestDB(t)
	defer cleanup()

	ctx := context.Background()
	if err := createEbayFixture(ctx, dbClient, "query-1", "https://example.test/one", 100); err != nil {
		t.Fatalf("Failed to create first fixture: %v", err)
	}

	first, err := dbClient.ListSavedQueries(ctx, "ebay")
	if err != nil {
		t.Fatalf("ListSavedQueries failed: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("Expected 1 query, got %d", len(first))
	}

	mutatedUrl := "https://example.test/mutated"
	first[0].Url = &mutatedUrl

	if err := createEbayFixture(ctx, dbClient, "query-2", "https://example.test/two", 200); err != nil {
		t.Fatalf("Failed to create second fixture: %v", err)
	}

	second, err := dbClient.ListSavedQueries(ctx, "ebay")
	if err != nil {
		t.Fatalf("ListSavedQueries failed: %v", err)
	}
	if len(second) != 1 {
		t.Fatalf("Expected cached query count to remain 1, got %d", len(second))
	}
	if second[0].Url == nil || *second[0].Url != "https://example.test/one" {
		t.Fatalf("Expected cached URL to remain unchanged, got %#v", second[0].Url)
	}
}

func TestQueryWritesInvalidateSavedQueryCache(t *testing.T) {
	dbClient, cleanup := setupQueryCacheTestDB(t)
	defer cleanup()

	ctx := context.Background()
	if err := createEbayFixture(ctx, dbClient, "query-1", "https://example.test/one", 100); err != nil {
		t.Fatalf("Failed to create first fixture: %v", err)
	}

	cached, err := dbClient.ListSavedQueries(ctx, "ebay")
	if err != nil {
		t.Fatalf("ListSavedQueries failed: %v", err)
	}
	if len(cached) != 1 {
		t.Fatalf("Expected 1 cached query, got %d", len(cached))
	}

	created, err := dbClient.CreateEbayQuery(ctx, false, &models.Ebay{
		Url:      "https://example.test/two",
		MaxPrice: 200,
	})
	if err != nil {
		t.Fatalf("CreateEbayQuery failed: %v", err)
	}
	if created.Url == nil || *created.Url != "https://example.test/two" {
		t.Fatalf("Expected created query URL, got %#v", created.Url)
	}

	refreshed, err := dbClient.ListSavedQueries(ctx, "ebay")
	if err != nil {
		t.Fatalf("ListSavedQueries failed: %v", err)
	}
	if len(refreshed) != 2 {
		t.Fatalf("Expected cache to refresh to 2 queries, got %d", len(refreshed))
	}
}

func createEbayFixture(ctx context.Context, dbClient *Query, queryId string, url string, maxPrice float64) error {
	return dbClient.Transaction(func(tx *Query) error {
		if err := tx.SearchQuery.WithContext(ctx).Create(&models.SearchQuery{
			ID:        queryId,
			DmOnly:    false,
			CreatedAt: time.Now().UTC(),
		}); err != nil {
			return err
		}

		return tx.Ebay.WithContext(ctx).Create(&models.Ebay{
			Url:      url,
			MaxPrice: maxPrice,
			QueryId:  queryId,
		})
	})
}
