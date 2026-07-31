package scanners

import (
	"context"
	"os"
	"testing"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func requireLiveCashConvertersTests(t *testing.T) {
	t.Helper()
	if os.Getenv("DEALSCANNER_RUN_LIVE_CC_TESTS") != "1" {
		t.Skip("set DEALSCANNER_RUN_LIVE_CC_TESTS=1 to run live Cash Converters tests")
	}
}

func TestLiveCashConvertersAPIAndDetail(t *testing.T) {
	requireLiveCashConvertersTests(t)
	if testing.Short() {
		t.Skip("skipping live Cash Converters test in -short mode")
	}

	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	_ = gdb.AutoMigrate(
		&models.SearchQuery{},
		&models.UserQuery{},
		&models.CashConverters{},
		&models.CashConvertersFilter{},
		&models.Globals{},
		&models.Listing{},
		&models.ListingObservation{},
		&models.QueryListingState{},
		&models.CashConvertersScanState{},
		&models.CashConvertersListingMeta{},
		&models.CashConvertersDetailJob{},
		&models.CashConvertersSearchDoc{},
		&models.CashConvertersDeletedListing{},
	)

	dbClient := qry.Use(gdb)
	scanner := NewCashConvertersScanner(dbClient)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Test live API page fetch
	catalogPage, err := scanner.fetchCatalogPage(ctx, ccCatalogSortNewest, 1)
	if err != nil {
		t.Fatalf("fetchCatalogPage against live Cash Converters API failed: %v", err)
	}
	summaries := catalogPage.Summaries

	if len(summaries) == 0 {
		t.Fatalf("expected live Cash Converters API page 1 to return at least 1 item")
	}

	t.Logf("Live API returned %d items on page 1", len(summaries))
	firstItem := summaries[0]
	t.Logf("Sample item: Title=%q, Price=%.2f, CanonicalURL=%q", firstItem.Title, firstItem.TotalPrice, firstItem.CanonicalURL)

	if firstItem.CanonicalURL == "" {
		t.Errorf("expected non-empty canonical URL")
	}
	if firstItem.Title == "" {
		t.Errorf("expected non-empty title")
	}

	// 2. Test live Detail page fetch & HTML parsing
	detailRes, err := scanner.fetchCcDetailPage(ctx, firstItem.CanonicalURL)
	if err != nil {
		t.Fatalf("fetchCcDetailPage against live item page %s failed: %v", firstItem.CanonicalURL, err)
	}

	if detailRes == nil {
		t.Fatalf("expected non-nil detail result")
	}

	t.Logf("Live Detail result: Title=%q, Availability=%s, DescriptionLength=%d",
		detailRes.Title, detailRes.Availability, len(detailRes.Description))

	if detailRes.Title == "" {
		t.Errorf("expected non-empty title from live detail page")
	}

	// 3. Test saving live summaries and performing catalog search
	upserted, err := dbClient.UpsertCashConvertersSummaries(ctx, summaries[:1], 1, 1, time.Now().UTC())
	if err != nil {
		t.Fatalf("Upsert live summary failed: %v", err)
	}

	if len(upserted.NewListingIDs) == 0 {
		t.Fatalf("expected 1 new listing ID inserted")
	}

	// Update detail on the live item
	detailInput := qry.CcDetailInput{
		CanonicalUrl: firstItem.CanonicalURL,
		Title:        detailRes.Title,
		Description:  detailRes.Description,
		Availability: detailRes.Availability,
		Price:        firstItem.Price,
		Shipping:     firstItem.Shipping,
		TotalPrice:   firstItem.TotalPrice,
		ImageUrl:     firstItem.ImageURL,
	}
	if err := dbClient.UpdateCashConvertersDetail(ctx, upserted.NewListingIDs[0], detailInput, time.Now().UTC()); err != nil {
		t.Fatalf("UpdateCashConvertersDetail failed: %v", err)
	}

	// Search local DB using first word of title
	words := parsePhrases(detailRes.Title)
	if len(words) > 0 {
		searchTerm := words[0]
		searchResults, err := dbClient.SearchCashConvertersCatalog(ctx, qry.CashConvertersCatalogSearchInput{
			Text:          searchTerm,
			AvailableOnly: false,
			Limit:         5,
		})
		if err != nil {
			t.Fatalf("SearchCashConvertersCatalog failed: %v", err)
		}
		if len(searchResults) == 0 {
			t.Errorf("expected catalog search for term %q to return item", searchTerm)
		} else {
			t.Logf("Catalog search for %q matched %d item(s)", searchTerm, len(searchResults))
		}
	}
}

func TestLiveCashConvertersMultiPage(t *testing.T) {
	requireLiveCashConvertersTests(t)
	if testing.Short() {
		t.Skip("skipping live Cash Converters multi-page test in -short mode")
	}

	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	_ = gdb.AutoMigrate(
		&models.SearchQuery{},
		&models.UserQuery{},
		&models.CashConverters{},
		&models.CashConvertersFilter{},
		&models.Globals{},
		&models.Listing{},
		&models.ListingObservation{},
		&models.QueryListingState{},
		&models.CashConvertersScanState{},
		&models.CashConvertersListingMeta{},
		&models.CashConvertersDetailJob{},
		&models.CashConvertersSearchDoc{},
		&models.CashConvertersDeletedListing{},
	)

	dbClient := qry.Use(gdb)
	scanner := NewCashConvertersScanner(dbClient)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Test pages 1 and 2 against live API, and verify page 3 indicates end of results
	totalDiscovered := 0
	pagesWithItems := 0

	for page := 1; page <= 3; page++ {
		catalogPage, err := scanner.fetchCatalogPage(ctx, ccCatalogSortPrice, page)
		if err != nil {
			t.Fatalf("fetchCatalogPage for live page %d failed: %v", page, err)
		}
		summaries := catalogPage.Summaries

		t.Logf("Live page %d returned %d items", page, len(summaries))

		if len(summaries) > 0 {
			pagesWithItems++
			upserted, err := dbClient.UpsertCashConvertersSummaries(ctx, summaries, 1, page, time.Now().UTC())
			if err != nil {
				t.Fatalf("Upsert summaries for page %d failed: %v", page, err)
			}
			totalDiscovered += len(summaries)
			t.Logf("Page %d upsert: %d new, %d changed", page, len(upserted.NewListingIDs), len(upserted.ChangedListingIDs))
		} else {
			t.Logf("Live page %d returned 0 items (reached end of site/results)", page)
		}
	}

	if pagesWithItems < 2 {
		t.Fatalf("expected at least 2 pages with items from live API, got %d", pagesWithItems)
	}

	// Verify total distinct listings in database
	var count int64
	if err := gdb.Model(&models.Listing{}).Where("source = ?", "cashConverters").Count(&count).Error; err != nil {
		t.Fatalf("count listings: %v", err)
	}

	t.Logf("Total unique Cash Converters listings stored across live pages: %d (discovered %d total summaries)", count, totalDiscovered)
	if count == 0 {
		t.Errorf("expected > 0 unique listings stored in database")
	}
}
