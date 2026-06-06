package db

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = gdb.AutoMigrate(&Query{}, &Listing{}, &ListingObservation{}, &QueryListingState{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbClient := &DB{gdb}

	cleanup := func() {
		sqlDB, _ := gdb.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return dbClient, cleanup
}

func TestParsePhraseList(t *testing.T) {
	parsed := ParsePhraseList(" xbox, Series X ,, Controller ")
	expected := []string{"xbox", "series x", "controller"}
	if len(parsed) != len(expected) {
		t.Fatalf("Expected %d elements, got %d", len(expected), len(parsed))
	}
	for i, v := range parsed {
		if v != expected[i] {
			t.Errorf("Expected element %d to be %q, got %q", i, expected[i], v)
		}
	}
}

func TestMatchPhrases(t *testing.T) {
	res := MatchPhrases("Xbox console\nIncludes an elite controller", "xbox, elite controller", "")
	if res.Type != MatchTypeMatched {
		t.Errorf("Expected MatchTypeMatched, got %v", res.Type)
	}

	res = MatchPhrases("Xbox console", "playstation", "")
	if res.Type != MatchTypeRejected || res.Reason != "MissingRequiredPhrase" {
		t.Errorf("Expected rejected MissingRequiredPhrase, got type %v, reason %q", res.Type, res.Reason)
	}

	res = MatchPhrases("Xbox console damaged", "xbox", "damaged")
	if res.Type != MatchTypeRejected || res.Reason != "ExcludedPhrase" {
		t.Errorf("Expected rejected ExcludedPhrase, got type %v, reason %q", res.Type, res.Reason)
	}
}

func TestMatchPriceRange(t *testing.T) {
	maxPrice := 100.0
	res := MatchPriceRange(120.0, nil, &maxPrice)
	if res.Type != MatchTypeRejected || res.Reason != "AboveMaxPrice" {
		t.Errorf("Expected rejected AboveMaxPrice, got type %v, reason %q", res.Type, res.Reason)
	}

	minPrice := 100.0
	res = MatchPriceRange(80.0, &minPrice, nil)
	if res.Type != MatchTypeRejected || res.Reason != "BelowMinPrice" {
		t.Errorf("Expected rejected BelowMinPrice, got type %v, reason %q", res.Type, res.Reason)
	}

	minPrice = 80.0
	maxPrice = 120.0
	res = MatchPriceRange(100.0, &minPrice, &maxPrice)
	if res.Type != MatchTypeMatched {
		t.Errorf("Expected matched, got %v", res.Type)
	}
}

func TestCombineMatchResults(t *testing.T) {
	phraseResult := MatchResult{Type: MatchTypeRejected, Reason: "ExcludedPhrase"}
	priceResult := MatchResult{Type: MatchTypeRejected, Reason: "AboveMaxPrice"}
	res := CombineMatchResults(phraseResult, priceResult)
	if res.Type != MatchTypeRejected || res.Reason != "ExcludedPhrase" {
		t.Errorf("Expected ExcludedPhrase, got type %v, reason %q", res.Type, res.Reason)
	}

	res = CombineMatchResults(MatchResult{Type: MatchTypeMatched}, MatchResult{Type: MatchTypeMatched})
	if res.Type != MatchTypeMatched {
		t.Errorf("Expected Matched, got %v", res.Type)
	}
}

func TestEvaluateListingForQuery(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	query := Query{ID: "test-query", CreatedAt: time.Now()}
	if err := dbClient.Create(&query).Error; err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	canonicalUrl := "https://example.test/listing-123"
	listingId := StableListingId("ebay", canonicalUrl)

	firstListing := DiscoveredListing{
		Source:       "ebay",
		CanonicalUrl: canonicalUrl,
		Title:        "Price drop listing",
		Price:        float64Ptr(150.0),
		TotalPrice:   float64Ptr(150.0),
		Availability: stringPtr("available"),
	}

	obsId1, err := dbClient.PersistListingObservation(ctx, firstListing, time.Now())
	if err != nil || obsId1 == "" {
		t.Fatalf("Failed to persist listing observation: %v", err)
	}

	maxPrice := 100.0
	match1 := MatchPriceRange(*firstListing.TotalPrice, nil, &maxPrice)

	dec1, err := dbClient.EvaluateListingForQuery(ctx, query.ID, listingId, "ebay", firstListing.TotalPrice, match1, time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}
	if dec1.Type != DecisionTypeDoNotNotify || dec1.Reason != DecisionReasonRejected {
		t.Errorf("Expected DoNotNotify Rejected, got type %q, reason %q", dec1.Type, dec1.Reason)
	}

	secondListing := firstListing
	secondListing.Price = float64Ptr(90.0)
	secondListing.TotalPrice = float64Ptr(90.0)

	obsId2, err := dbClient.PersistListingObservation(ctx, secondListing, time.Now())
	if err != nil || obsId2 == "" {
		t.Fatalf("Failed to persist listing observation: %v", err)
	}

	match2 := MatchPriceRange(*secondListing.TotalPrice, nil, &maxPrice)

	dec2, err := dbClient.EvaluateListingForQuery(ctx, query.ID, listingId, "ebay", secondListing.TotalPrice, match2, time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}
	if dec2.Type != DecisionTypeNotify || dec2.Reason != DecisionReasonPriceDroppedIntoRange {
		t.Errorf("Expected Notify PriceDroppedIntoRange, got type %q, reason %q", dec2.Type, dec2.Reason)
	}

	dec3, err := dbClient.EvaluateListingForQuery(ctx, query.ID, listingId, "ebay", secondListing.TotalPrice, match2, time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}
	if dec3.Type != DecisionTypeDoNotNotify || dec3.Reason != DecisionReasonAlreadyNotified {
		t.Errorf("Expected DoNotNotify AlreadyNotified, got type %q, reason %q", dec3.Type, dec3.Reason)
	}

	var count int64
	dbClient.Model(&ListingObservation{}).Where("listingId = ?", listingId).Count(&count)
	if count != 2 {
		t.Errorf("Expected 2 observations, got %d", count)
	}
}

func TestFurtherPriceDrops(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	query := Query{ID: "test-query-drop", CreatedAt: time.Now()}
	dbClient.Create(&query)

	canonicalUrl := "https://example.test/further-drop-123"
	listingId := StableListingId("ebay", canonicalUrl)

	firstListing := DiscoveredListing{
		Source:       "ebay",
		CanonicalUrl: canonicalUrl,
		Title:        "Further price drop listing",
		Price:        float64Ptr(100.0),
		TotalPrice:   float64Ptr(100.0),
		Availability: stringPtr("available"),
	}

	dbClient.PersistListingObservation(ctx, firstListing, time.Now())
	match := MatchResult{Type: MatchTypeMatched}

	dec1, err := dbClient.EvaluateListingForQuery(ctx, query.ID, listingId, "ebay", firstListing.TotalPrice, match, time.Now(), 0.05)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}
	if dec1.Type != DecisionTypeNotify || dec1.Reason != DecisionReasonFirstMatch {
		t.Errorf("Expected Notify FirstMatch, got type %q, reason %q", dec1.Type, dec1.Reason)
	}

	secondListing := firstListing
	secondListing.Price = float64Ptr(94.0)
	secondListing.TotalPrice = float64Ptr(94.0)

	dbClient.PersistListingObservation(ctx, secondListing, time.Now())
	dec2, err := dbClient.EvaluateListingForQuery(ctx, query.ID, listingId, "ebay", secondListing.TotalPrice, match, time.Now(), 0.05)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}
	if dec2.Type != DecisionTypeNotify || dec2.Reason != DecisionReasonFurtherPriceDrop {
		t.Errorf("Expected Notify FurtherPriceDrop, got type %q, reason %q", dec2.Type, dec2.Reason)
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}
