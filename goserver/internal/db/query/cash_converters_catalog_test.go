package query

import (
	"context"
	"fmt"
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

func TestCatalogSightingsDoNotRestoreDetailConfirmedSoldListing(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	summary := CashConvertersDiscoveredSummary{
		CanonicalURL: "https://www.cashconverters.com.au/shop/item/sold",
		Title:        "Gaming PC",
		TotalPrice:   999,
	}

	created, err := q.UpsertCashConvertersSummaries(ctx, []CashConvertersDiscoveredSummary{summary}, 1, 1, now)
	if err != nil {
		t.Fatalf("create summary: %v", err)
	}
	listingID := created.NewListingIDs[0]
	if err := q.UpdateCashConvertersDetail(ctx, listingID, CcDetailInput{
		CanonicalUrl: summary.CanonicalURL,
		Title:        summary.Title,
		Availability: "unavailable",
	}, now.Add(time.Minute)); err != nil {
		t.Fatalf("mark detail unavailable: %v", err)
	}

	if _, err := q.UpsertCashConvertersSummaries(ctx, []CashConvertersDiscoveredSummary{summary}, 1, 1, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("re-upsert sold listing: %v", err)
	}

	var listing models.Listing
	if err := q.db.Where("id = ?", listingID).First(&listing).Error; err != nil {
		t.Fatalf("load listing: %v", err)
	}
	if listing.Availability == nil || *listing.Availability != "unavailable" {
		t.Fatalf("catalogue incorrectly restored listing availability: %v", listing.Availability)
	}
	var doc models.CashConvertersSearchDoc
	if err := q.db.Where("listingId = ?", listingID).First(&doc).Error; err != nil {
		t.Fatalf("load search doc: %v", err)
	}
	if doc.Availability != "unavailable" {
		t.Fatalf("catalogue incorrectly restored search availability: %q", doc.Availability)
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

func TestUpsertCashConvertersSummariesRepairsOrphanedMetadata(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	canonicalURL := "https://www.cashconverters.com.au/shop/item/028800272284"
	listingID := "cashConverters:" + canonicalURL
	oldSeenAt := now.Add(-24 * time.Hour)
	missingAt := now.Add(-time.Hour)

	orphanedMeta := models.CashConvertersListingMeta{
		ListingID:          listingID,
		FirstSeenSweepID:   3,
		LastSeenSweepID:    3,
		LastSeenPage:       12,
		LastListedAt:       oldSeenAt,
		SourceStatus:       "confirmed_missing",
		MissingSuspectedAt: &missingAt,
		MissingVerifiedAt:  &missingAt,
		DeletedAt:          &missingAt,
		CreatedAt:          oldSeenAt,
		UpdatedAt:          missingAt,
	}
	if err := q.db.Create(&orphanedMeta).Error; err != nil {
		t.Fatalf("create orphaned metadata: %v", err)
	}

	summaries := []CashConvertersDiscoveredSummary{{
		CanonicalURL: canonicalURL,
		Title:        "Gold Bangle",
		Price:        150,
		TotalPrice:   165,
		ImageURL:     "https://example.com/bangle.jpg",
	}}
	result, err := q.UpsertCashConvertersSummaries(ctx, summaries, 9, 2, now)
	if err != nil {
		t.Fatalf("UpsertCashConvertersSummaries: %v", err)
	}
	if len(result.NewListingIDs) != 1 || result.NewListingIDs[0] != listingID {
		t.Fatalf("expected repaired listing %q to be new, got %v", listingID, result.NewListingIDs)
	}

	var repairedMeta models.CashConvertersListingMeta
	if err := q.db.Where("listingId = ?", listingID).First(&repairedMeta).Error; err != nil {
		t.Fatalf("load repaired metadata: %v", err)
	}
	if repairedMeta.SourceStatus != "available" || repairedMeta.LastSeenSweepID != 9 || repairedMeta.LastSeenPage != 2 {
		t.Errorf("metadata was not refreshed: %+v", repairedMeta)
	}
	if repairedMeta.MissingSuspectedAt != nil || repairedMeta.MissingVerifiedAt != nil || repairedMeta.DeletedAt != nil {
		t.Errorf("expected missing/deleted markers to be cleared: %+v", repairedMeta)
	}
	if repairedMeta.FirstSeenSweepID != 3 || !repairedMeta.CreatedAt.Equal(oldSeenAt) {
		t.Errorf("expected original first-seen metadata to be preserved: %+v", repairedMeta)
	}
}

func TestEnqueueAndProcessDetailJobs(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	listingID := "cashConverters:https://www.cashconverters.com.au/shop/item/123"

	if err := q.EnqueueCashConvertersDetailJobs(ctx, listingID, now, 10); err != nil {
		t.Fatalf("EnqueueCashConvertersDetailJobs: %v", err)
	}

	// Test idempotency
	if err := q.EnqueueCashConvertersDetailJobs(ctx, listingID, now, 10); err != nil {
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

func TestDueDetailJobsPrioritizeInitialCatalogEnrichment(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	startedAt := time.Now().UTC().Add(-6 * time.Hour)

	if err := q.EnqueueCashConvertersDetailJobs(ctx, "cashConverters:https://example.com/old", startedAt, 10); err != nil {
		t.Fatalf("enqueue old jobs: %v", err)
	}
	oldInitial, err := q.ListDueCashConvertersDetailJobs(ctx, 1, startedAt.Add(time.Second))
	if err != nil || len(oldInitial) != 1 {
		t.Fatalf("load old initial job: jobs=%d err=%v", len(oldInitial), err)
	}
	if err := q.CompleteCashConvertersDetailJob(ctx, oldInitial[0].ID, startedAt.Add(time.Second)); err != nil {
		t.Fatalf("complete old initial job: %v", err)
	}

	newListingID := "cashConverters:https://example.com/new"
	if err := q.EnqueueCashConvertersDetailJobs(ctx, newListingID, startedAt.Add(5*time.Hour), 100); err != nil {
		t.Fatalf("enqueue new jobs: %v", err)
	}
	due, err := q.ListDueCashConvertersDetailJobs(ctx, 1, time.Now().UTC())
	if err != nil {
		t.Fatalf("list prioritized jobs: %v", err)
	}
	if len(due) != 1 || due[0].ListingID != newListingID || due[0].Stage != "initial" {
		t.Fatalf("expected new initial job before overdue refresh jobs, got %+v", due)
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

	if err := q.db.Exec(`CREATE VIRTUAL TABLE CashConvertersSearchFts USING fts5(
		listingId UNINDEXED, title, description, canonicalUrl UNINDEXED,
		content='CashConvertersSearchDoc', content_rowid='rowid'
	)`).Error; err != nil {
		t.Skipf("SQLite test build does not include FTS5: %v", err)
	}
	for _, statement := range []string{
		`CREATE TRIGGER cc_search_doc_ai AFTER INSERT ON CashConvertersSearchDoc BEGIN
			INSERT INTO CashConvertersSearchFts(rowid, listingId, title, description, canonicalUrl)
			VALUES (new.rowid, new.listingId, new.title, new.description, new.canonicalUrl); END`,
		`CREATE TRIGGER cc_search_doc_ad AFTER DELETE ON CashConvertersSearchDoc BEGIN
			INSERT INTO CashConvertersSearchFts(CashConvertersSearchFts, rowid, listingId, title, description, canonicalUrl)
			VALUES ('delete', old.rowid, old.listingId, old.title, old.description, old.canonicalUrl); END`,
		`CREATE TRIGGER cc_search_doc_au AFTER UPDATE ON CashConvertersSearchDoc BEGIN
			INSERT INTO CashConvertersSearchFts(CashConvertersSearchFts, rowid, listingId, title, description, canonicalUrl)
			VALUES ('delete', old.rowid, old.listingId, old.title, old.description, old.canonicalUrl);
			INSERT INTO CashConvertersSearchFts(rowid, listingId, title, description, canonicalUrl)
			VALUES (new.rowid, new.listingId, new.title, new.description, new.canonicalUrl); END`,
	} {
		if err := q.db.Exec(statement).Error; err != nil {
			t.Fatalf("create Cash Converters FTS trigger: %v", err)
		}
	}
	if err := q.db.Exec("INSERT INTO CashConvertersSearchFts(CashConvertersSearchFts) VALUES ('rebuild')").Error; err != nil {
		t.Fatalf("rebuild Cash Converters FTS index: %v", err)
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

func TestCashConvertersBatchDiscoverySchedulesOnlyNewListings(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	var summaries []CashConvertersDiscoveredSummary
	for index := range 81 {
		summaries = append(summaries, CashConvertersDiscoveredSummary{
			CanonicalURL: fmt.Sprintf("https://example.com/item/%d", index),
			Title:        fmt.Sprintf("Item %d", index), TotalPrice: float64(index + 1),
		})
	}
	// A duplicate must not create another observation or another job schedule.
	summaries = append(summaries, CashConvertersDiscoveredSummary{
		CanonicalURL: "https://example.com/item/0", Title: "Item 0", TotalPrice: 1,
	})
	reads, jobWrites := 0, 0
	if err := q.db.Callback().Query().After("gorm:query").Register("test:count_catalog_reads", func(db *gorm.DB) {
		if db.Statement.Table == "Listing" || db.Statement.Table == "CashConvertersSearchDoc" {
			reads++
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.db.Callback().Create().After("gorm:create").Register("test:count_job_writes", func(db *gorm.DB) {
		if db.Statement.Table == "CashConvertersDetailJob" {
			jobWrites++
		}
	}); err != nil {
		t.Fatal(err)
	}
	created, err := q.UpsertCashConvertersSummariesWithDetailJobs(ctx, summaries, 1, 1, now, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(created.NewListingIDs) != 81 || len(created.ListingIDs) != 81 || reads != 6 {
		t.Fatalf("expected 81 unique listings and two reads per batch, got new=%d ids=%d reads=%d", len(created.NewListingIDs), len(created.ListingIDs), reads)
	}
	if jobWrites != 11 {
		t.Fatalf("expected 11 job batches for 405 jobs, got %d", jobWrites)
	}
	reads, jobWrites = 0, 0
	seenAt := now.Add(time.Minute)
	repeated, err := q.UpsertCashConvertersSummariesWithDetailJobs(ctx, summaries, 2, 3, seenAt, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(repeated.NewListingIDs) != 0 || len(repeated.ChangedListingIDs) != 0 || !repeated.HadPreviouslySeen {
		t.Fatalf("unchanged sightings reported changes: %+v", repeated)
	}
	if reads != 6 || jobWrites != 0 {
		t.Fatalf("expected six batch reads and no job inserts on repeat, got reads=%d jobWrites=%d", reads, jobWrites)
	}
	var observations, jobs int64
	if err := q.db.Model(&models.ListingObservation{}).Count(&observations).Error; err != nil {
		t.Fatal(err)
	}
	if err := q.db.Model(&models.CashConvertersDetailJob{}).Count(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	if observations != 81 || jobs != 405 {
		t.Fatalf("unexpected observations=%d jobs=%d", observations, jobs)
	}
	var doc models.CashConvertersSearchDoc
	if err := q.db.Where("listingId = ?", created.NewListingIDs[0]).First(&doc).Error; err != nil {
		t.Fatal(err)
	}
	if !doc.LastSeenAt.Equal(seenAt) || !doc.UpdatedAt.Equal(now) {
		t.Fatalf("freshness should advance without changing content timestamp: %+v", doc)
	}
	var meta models.CashConvertersListingMeta
	if err := q.db.Where("listingId = ?", doc.ListingID).First(&meta).Error; err != nil {
		t.Fatal(err)
	}
	if meta.LastSeenSweepID != 2 || meta.LastSeenPage != 3 || !meta.LastListedAt.Equal(seenAt) {
		t.Fatalf("sweep freshness was lost: %+v", meta)
	}
	var initial models.CashConvertersDetailJob
	if err := q.db.Where("listingId = ? AND stage = ?", doc.ListingID, "initial").First(&initial).Error; err != nil {
		t.Fatal(err)
	}
	if !initial.DueAt.Equal(now) || initial.Priority != 100 {
		t.Fatalf("repeat sighting changed original job schedule: %+v", initial)
	}
}

func TestCashConvertersBatchDiscoveryRollsBackFailedJobScheduling(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	if err := q.db.Exec(`CREATE TRIGGER reject_detail_job BEFORE INSERT ON CashConvertersDetailJob
		BEGIN SELECT RAISE(ABORT, 'injected job failure'); END`).Error; err != nil {
		t.Fatal(err)
	}
	result, err := q.UpsertCashConvertersSummariesWithDetailJobs(context.Background(), []CashConvertersDiscoveredSummary{{
		CanonicalURL: "https://example.com/item/rollback", Title: "Rollback", TotalPrice: 10,
	}}, 1, 1, time.Now().UTC(), 100)
	if err == nil || len(result.ListingIDs) != 0 {
		t.Fatalf("expected an error and no committed results: result=%+v err=%v", result, err)
	}
	for _, table := range []string{"Listing", "ListingObservation", "CashConvertersListingMeta", "CashConvertersSearchDoc", "CashConvertersDetailJob"} {
		var count int64
		if err := q.db.Table(table).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s retained %d rows after rollback", table, count)
		}
	}
}

func TestCashConvertersBatchPreservesDetailsAndRepairsMissingChildren(t *testing.T) {
	q := setupCcCatalogTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	extID := "123456"
	summary := CashConvertersDiscoveredSummary{CanonicalURL: "https://example.com/item/repair", Title: "Console", ExternalItemID: &extID, TotalPrice: 200}
	result, err := q.UpsertCashConvertersSummaries(ctx, []CashConvertersDiscoveredSummary{summary}, 1, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	id := result.NewListingIDs[0]
	fetchedAt := now.Add(time.Minute)
	if err := q.UpdateCashConvertersDetail(ctx, id, CcDetailInput{Title: "Console", Description: "Detailed description", Availability: "unavailable"}, fetchedAt); err != nil {
		t.Fatal(err)
	}
	summary.ExternalItemID = nil
	summary.Title = "Updated console"
	summary.TotalPrice = 150
	for _, repair := range []bool{false, true} {
		if repair {
			if err := q.db.Where("listingId = ?", id).Delete(&models.CashConvertersSearchDoc{}).Error; err != nil {
				t.Fatal(err)
			}
			if err := q.db.Where("listingId = ?", id).Delete(&models.CashConvertersListingMeta{}).Error; err != nil {
				t.Fatal(err)
			}
		}
		updated, err := q.UpsertCashConvertersSummaries(ctx, []CashConvertersDiscoveredSummary{summary}, 2, 2, now.Add(2*time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		if len(updated.NewListingIDs) != 0 || len(updated.ChangedListingIDs) != 1 {
			t.Fatalf("unexpected changed result: %+v", updated)
		}
		var doc models.CashConvertersSearchDoc
		if err := q.db.Where("listingId = ?", id).First(&doc).Error; err != nil {
			t.Fatal(err)
		}
		if doc.Description != "Detailed description" || doc.Availability != "unavailable" || doc.LastDetailAt == nil || !doc.LastDetailAt.Equal(fetchedAt) || doc.TotalPrice == nil || *doc.TotalPrice != 150 {
			t.Fatalf("batch lost detail or price data: %+v", doc)
		}
		var meta models.CashConvertersListingMeta
		if err := q.db.Where("listingId = ?", id).First(&meta).Error; err != nil {
			t.Fatal(err)
		}
		if meta.ExternalItemID == nil || *meta.ExternalItemID != extID || meta.LastSeenSweepID != 2 {
			t.Fatalf("batch lost metadata: %+v", meta)
		}
	}
}
