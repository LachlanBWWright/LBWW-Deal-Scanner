package query

import (
	"context"
	"testing"
	"time"

	"dealscanner/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCcCatalogTestDB(t *testing.T) *Query {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	err = gdb.AutoMigrate(
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
	if err != nil {
		t.Fatalf("auto-migrate test tables: %v", err)
	}
	return Use(gdb)
}

func TestCashConvertersScanState(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()

	state, err := q.GetCashConvertersScanState(ctx)
	if err != nil {
		t.Fatalf("GetCashConvertersScanState: %v", err)
	}
	if state.ID != "cashConverters" {
		t.Errorf("expected ID 'cashConverters', got %s", state.ID)
	}
	if state.CurrentSweepID != 1 {
		t.Errorf("expected sweep ID 1, got %d", state.CurrentSweepID)
	}
	if state.TailNextPage != 1 {
		t.Errorf("expected tail next page 1, got %d", state.TailNextPage)
	}

	state.TailNextPage = 5
	state.CurrentSweepID = 2
	if err := q.UpdateCashConvertersScanState(ctx, state); err != nil {
		t.Fatalf("UpdateCashConvertersScanState: %v", err)
	}

	stateReloaded, err := q.GetCashConvertersScanState(ctx)
	if err != nil {
		t.Fatalf("GetCashConvertersScanState reloaded: %v", err)
	}
	if stateReloaded.TailNextPage != 5 || stateReloaded.CurrentSweepID != 2 {
		t.Errorf("expected page 5 sweep 2, got page %d sweep %d", stateReloaded.TailNextPage, stateReloaded.CurrentSweepID)
	}
}

func TestUpsertCashConvertersSummaries(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	extID := "023500607432"
	summaries := []CashConvertersDiscoveredSummary{
		{
			CanonicalURL:   "https://www.cashconverters.com.au/shop/item/023500607432",
			ExternalItemID: &extID,
			Title:          "PlayStation 5 Console",
			Price:          500.0,
			Shipping:       15.0,
			TotalPrice:     515.0,
			ImageURL:       "https://example.com/ps5.jpg",
		},
	}

	upserted, err := q.UpsertCashConvertersSummaries(ctx, summaries, 1, 1, now)
	if err != nil {
		t.Fatalf("UpsertCashConvertersSummaries: %v", err)
	}

	if len(upserted.NewListingIDs) != 1 {
		t.Fatalf("expected 1 new listing ID, got %d", len(upserted.NewListingIDs))
	}
	if len(upserted.ChangedListingIDs) != 1 {
		t.Fatalf("expected 1 changed listing ID, got %d", len(upserted.ChangedListingIDs))
	}

	listingID := upserted.NewListingIDs[0]

	var lst models.Listing
	if err := q.db.Where("id = ?", listingID).First(&lst).Error; err != nil {
		t.Fatalf("failed to find created listing: %v", err)
	}
	if lst.Title != "PlayStation 5 Console" {
		t.Errorf("expected title 'PlayStation 5 Console', got %s", lst.Title)
	}

	var meta models.CashConvertersListingMeta
	if err := q.db.Where("listingId = ?", listingID).First(&meta).Error; err != nil {
		t.Fatalf("failed to find created metadata: %v", err)
	}
	if meta.SourceStatus != "available" {
		t.Errorf("expected status 'available', got %s", meta.SourceStatus)
	}

	// Re-upsert with price change
	summaries[0].Price = 450.0
	summaries[0].TotalPrice = 465.0
	upserted2, err := q.UpsertCashConvertersSummaries(ctx, summaries, 1, 1, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("Re-upsert failed: %v", err)
	}
	if len(upserted2.NewListingIDs) != 0 {
		t.Errorf("expected 0 new listing IDs on re-upsert, got %d", len(upserted2.NewListingIDs))
	}
	if len(upserted2.ChangedListingIDs) != 1 {
		t.Errorf("expected 1 changed listing ID on re-upsert, got %d", len(upserted2.ChangedListingIDs))
	}
	if !upserted2.HadPreviouslySeen {
		t.Error("expected re-upsert to report previously seen inventory")
	}

	upserted3, err := q.UpsertCashConvertersSummaries(ctx, summaries, 1, 1, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("unchanged re-upsert failed: %v", err)
	}
	if len(upserted3.ChangedListingIDs) != 0 {
		t.Errorf("expected unchanged listing not to be re-evaluated, got %d changed IDs", len(upserted3.ChangedListingIDs))
	}
}

func TestUpsertCashConvertersSummariesRepairsOrphanedSearchDoc(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	canonicalURL := "https://www.cashconverters.com.au/shop/item/050100270259"
	listingID := "cashConverters:" + canonicalURL

	orphanedDoc := models.CashConvertersSearchDoc{
		ListingID:    listingID,
		Title:        "Old title",
		Description:  "Preserve detailed description",
		CanonicalURL: canonicalURL,
		Availability: "available",
		SourceStatus: "available",
		LastSeenAt:   now.Add(-time.Hour),
		UpdatedAt:    now.Add(-time.Hour),
	}
	if err := q.db.Create(&orphanedDoc).Error; err != nil {
		t.Fatalf("create orphaned search doc: %v", err)
	}

	summaries := []CashConvertersDiscoveredSummary{{
		CanonicalURL: canonicalURL,
		Title:        "Current title",
		Price:        400,
		TotalPrice:   420,
		ImageURL:     "https://example.com/current.jpg",
	}}
	result, err := q.UpsertCashConvertersSummaries(ctx, summaries, 1, 1, now)
	if err != nil {
		t.Fatalf("UpsertCashConvertersSummaries: %v", err)
	}
	if len(result.NewListingIDs) != 1 {
		t.Fatalf("expected repaired listing to be new, got %d new IDs", len(result.NewListingIDs))
	}

	var repairedDoc models.CashConvertersSearchDoc
	if err := q.db.Where("listingId = ?", listingID).First(&repairedDoc).Error; err != nil {
		t.Fatalf("load repaired search doc: %v", err)
	}
	if repairedDoc.Title != "Current title" {
		t.Errorf("expected current title, got %q", repairedDoc.Title)
	}
	if repairedDoc.Description != "Preserve detailed description" {
		t.Errorf("expected existing description to be preserved, got %q", repairedDoc.Description)
	}
}

func TestEnqueueAndProcessDetailJobs(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	listingID := "cashConverters:https://www.cashconverters.com.au/shop/item/123"

	if err := q.EnqueueCashConvertersDetailJobs(ctx, listingID, now); err != nil {
		t.Fatalf("EnqueueCashConvertersDetailJobs: %v", err)
	}

	// Test idempotency
	if err := q.EnqueueCashConvertersDetailJobs(ctx, listingID, now); err != nil {
		t.Fatalf("Re-enqueue detail jobs failed: %v", err)
	}

	jobs, err := q.ListDueCashConvertersDetailJobs(ctx, 10, now.Add(time.Second))
	if err != nil {
		t.Fatalf("ListDueCashConvertersDetailJobs: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected exactly 1 initial detail job after idempotent enqueue, got %d", len(jobs))
	}

	job := jobs[0]
	if err := q.CompleteCashConvertersDetailJob(ctx, job.ID, now); err != nil {
		t.Fatalf("CompleteCashConvertersDetailJob: %v", err)
	}

	var reloadedJob models.CashConvertersDetailJob
	if err := q.db.Where("id = ?", job.ID).First(&reloadedJob).Error; err != nil {
		t.Fatalf("reload completed job: %v", err)
	}
	if reloadedJob.CompletedAt == nil {
		t.Errorf("expected job to be completed")
	}
}

func TestUpdateCashConvertersDetailAndSearch(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	extID := "023500607432"
	summaries := []CashConvertersDiscoveredSummary{
		{
			CanonicalURL:   "https://www.cashconverters.com.au/shop/item/023500607432",
			ExternalItemID: &extID,
			Title:          "Snapdragon X Laptop",
			Price:          600.0,
			Shipping:       0.0,
			TotalPrice:     600.0,
			ImageURL:       "https://example.com/laptop.jpg",
		},
	}

	upserted, err := q.UpsertCashConvertersSummaries(ctx, summaries, 1, 1, now)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	listingID := upserted.NewListingIDs[0]

	detail := CcDetailInput{
		CanonicalUrl: summaries[0].CanonicalURL,
		Title:        "Snapdragon X Laptop 16GB",
		Description:  "Ultra fast Qualcomm Snapdragon X Plus processor with 512GB SSD",
		Availability: "available",
		Price:        600.0,
		Shipping:     0.0,
		TotalPrice:   600.0,
		ImageUrl:     summaries[0].ImageURL,
	}

	if err := q.UpdateCashConvertersDetail(ctx, listingID, detail, now); err != nil {
		t.Fatalf("UpdateCashConvertersDetail: %v", err)
	}

	// Test Search
	results, err := q.SearchCashConvertersCatalog(ctx, CashConvertersCatalogSearchInput{
		Text:          "Snapdragon",
		AvailableOnly: true,
		Limit:         10,
	})
	if err != nil {
		t.Fatalf("SearchCashConvertersCatalog: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(results))
	}
	if results[0].ListingID != listingID {
		t.Errorf("expected listing ID %s, got %s", listingID, results[0].ListingID)
	}

	// Search description term
	resultsDesc, err := q.SearchCashConvertersCatalog(ctx, CashConvertersCatalogSearchInput{
		Text:          "Qualcomm",
		AvailableOnly: true,
		Limit:         10,
	})
	if err != nil {
		t.Fatalf("Search description term: %v", err)
	}
	if len(resultsDesc) != 1 {
		t.Fatalf("expected 1 result for description search, got %d", len(resultsDesc))
	}

	if err := q.db.Exec(`CREATE VIRTUAL TABLE CashConvertersSearchFts USING fts5(listingId UNINDEXED, title, description, canonicalUrl UNINDEXED)`).Error; err != nil {
		t.Skipf("SQLite test build does not include FTS5: %v", err)
	}
	if err := q.RefreshCashConvertersSearchDoc(ctx, listingID); err != nil {
		t.Fatalf("populate Cash Converters search index: %v", err)
	}

	filterCases := []struct {
		name  string
		input CashConvertersCatalogSearchInput
		want  int
	}{
		{
			name: "required whole word in description",
			input: CashConvertersCatalogSearchInput{
				Text: "Snapdragon", Required: "Qualcomm", RequiredMode: "all_words",
				AvailableOnly: true, Limit: 10,
			},
			want: 1,
		},
		{
			name: "required partial is not a whole word",
			input: CashConvertersCatalogSearchInput{
				Text: "Snapdragon", Required: "comm", RequiredMode: "all_words",
				AvailableOnly: true, Limit: 10,
			},
			want: 0,
		},
		{
			name: "required partial letter sequence",
			input: CashConvertersCatalogSearchInput{
				Text: "Snapdragon", Required: "comm", RequiredMode: "all_substring",
				AvailableOnly: true, Limit: 10,
			},
			want: 1,
		},
		{
			name: "excluded whole word in description",
			input: CashConvertersCatalogSearchInput{
				Text: "Snapdragon", Excluded: "Qualcomm", ExcludedMode: "any_words",
				AvailableOnly: true, Limit: 10,
			},
			want: 0,
		},
	}
	for _, tc := range filterCases {
		t.Run(tc.name, func(t *testing.T) {
			filtered, searchErr := q.SearchCashConvertersCatalog(ctx, tc.input)
			if searchErr != nil {
				t.Fatalf("SearchCashConvertersCatalog: %v", searchErr)
			}
			if len(filtered) != tc.want {
				t.Fatalf("result count = %d, want %d", len(filtered), tc.want)
			}
			count, countErr := q.CountCashConvertersCatalog(ctx, tc.input)
			if countErr != nil {
				t.Fatalf("CountCashConvertersCatalog: %v", countErr)
			}
			if count != int64(tc.want) {
				t.Fatalf("catalog count = %d, want %d", count, tc.want)
			}
		})
	}
}

func TestMissingAndPruningLifecycle(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	extID := "023500607999"
	summaries := []CashConvertersDiscoveredSummary{
		{
			CanonicalURL:   "https://www.cashconverters.com.au/shop/item/023500607999",
			ExternalItemID: &extID,
			Title:          "Vintage Camera",
			Price:          80.0,
			TotalPrice:     80.0,
			ImageURL:       "https://example.com/cam.jpg",
		},
	}

	upserted, err := q.UpsertCashConvertersSummaries(ctx, summaries, 1, 1, now)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	listingID := upserted.NewListingIDs[0]

	// Mark sweep 2 complete, where item was missing (sweep 1 item)
	if err := q.MarkCashConvertersMissingCandidates(ctx, 2, now); err != nil {
		t.Fatalf("MarkCashConvertersMissingCandidates: %v", err)
	}

	missingVerifs, err := q.ListCashConvertersMissingVerificationsDue(ctx, 10, now)
	if err != nil {
		t.Fatalf("ListCashConvertersMissingVerificationsDue: %v", err)
	}
	if len(missingVerifs) != 1 {
		t.Fatalf("expected 1 missing candidate, got %d", len(missingVerifs))
	}

	confirmed, err := q.RecordCashConvertersMissingVerification(ctx, listingID, now)
	if err != nil {
		t.Fatalf("first missing verification: %v", err)
	}
	if confirmed {
		t.Fatal("first non-404 missing verification confirmed the listing")
	}

	var meta models.CashConvertersListingMeta
	if err := q.db.Where("listingId = ?", listingID).First(&meta).Error; err != nil {
		t.Fatalf("reload metadata after first verification: %v", err)
	}
	if meta.MissingVerificationCount != 1 || meta.SourceStatus != "suspected_missing" {
		t.Fatalf("first verification left metadata at count=%d status=%q", meta.MissingVerificationCount, meta.SourceStatus)
	}

	if due, err := q.ListCashConvertersMissingVerificationsDue(ctx, 10, now.Add(time.Minute)); err != nil {
		t.Fatalf("list early retry: %v", err)
	} else if len(due) != 0 {
		t.Fatalf("expected verification retry to be delayed, got %d", len(due))
	}

	confirmed, err = q.RecordCashConvertersMissingVerification(ctx, listingID, now.Add(10*time.Minute))
	if err != nil {
		t.Fatalf("second missing verification: %v", err)
	}
	if !confirmed {
		t.Fatal("second non-404 missing verification did not confirm the listing")
	}

	// Prune
	pruned, err := q.DeleteConfirmedCashConvertersListings(ctx, now.Add(11*time.Minute), 10)
	if err != nil {
		t.Fatalf("DeleteConfirmedCashConvertersListings: %v", err)
	}
	if pruned != 1 {
		t.Errorf("expected 1 pruned listing, got %d", pruned)
	}

	// Check tombstone
	var tombstone models.CashConvertersDeletedListing
	if err := q.db.Where("listingId = ?", listingID).First(&tombstone).Error; err != nil {
		t.Errorf("expected tombstone record for %s", listingID)
	}
}
