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
	defaultGlobalsCache.invalidate()

	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = gdb.AutoMigrate(&models.SearchQuery{}, &models.Ebay{}, &models.Globals{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbClient := Use(gdb)
	cleanup := func() {
		defaultSavedQueryCache.invalidate()
		defaultGlobalsCache.invalidate()
		sqlDB, _ := gdb.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return dbClient, cleanup
}

func TestSavedQueryCacheDefensivelyCopiesLastPrice(t *testing.T) {
	price := 12.5
	item := QueryItem{LastPrice: &price}

	cloned := cloneQueryItem(item)
	if cloned.LastPrice == nil || *cloned.LastPrice != price {
		t.Fatalf("Expected cloned LastPrice %v, got %#v", price, cloned.LastPrice)
	}

	*cloned.LastPrice = 99
	if item.LastPrice == nil || *item.LastPrice != price {
		t.Fatalf("Expected original LastPrice to remain %v, got %#v", price, item.LastPrice)
	}
}

func TestGetGlobalsUsesCacheAndReturnsDefensiveCopies(t *testing.T) {
	dbClient, cleanup := setupQueryCacheTestDB(t)
	defer cleanup()

	ctx := context.Background()
	original := models.Globals{ID: "1", Gumtree: true}
	if err := dbClient.Globals.WithContext(ctx).Create(&original); err != nil {
		t.Fatalf("Failed to create globals: %v", err)
	}

	first, err := dbClient.GetGlobals(ctx)
	if err != nil {
		t.Fatalf("GetGlobals failed: %v", err)
	}
	if first == nil || !first.Gumtree {
		t.Fatalf("Expected cached Gumtree=true, got %#v", first)
	}
	first.Gumtree = false

	if _, err := dbClient.Globals.WithContext(ctx).Where(dbClient.Globals.ID.Eq("1")).Update(dbClient.Globals.Gumtree, false); err != nil {
		t.Fatalf("Failed to update globals directly: %v", err)
	}

	second, err := dbClient.GetGlobals(ctx)
	if err != nil {
		t.Fatalf("GetGlobals failed: %v", err)
	}
	if second == nil || !second.Gumtree {
		t.Fatalf("Expected cached defensive copy to keep Gumtree=true, got %#v", second)
	}

	second.Gumtree = false
	if err := dbClient.UpdateGlobals(ctx, second); err != nil {
		t.Fatalf("UpdateGlobals failed: %v", err)
	}

	updated, err := dbClient.GetGlobals(ctx)
	if err != nil {
		t.Fatalf("GetGlobals failed: %v", err)
	}
	if updated == nil || updated.Gumtree {
		t.Fatalf("Expected UpdateGlobals to refresh cache with Gumtree=false, got %#v", updated)
	}
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
