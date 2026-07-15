package query

import (
	"context"
	"testing"
	"time"

	"dealscanner/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*Query, func()) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = gdb.AutoMigrate(&models.SearchQuery{}, &models.Listing{}, &models.ListingObservation{}, &models.QueryListingState{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbClient := Use(gdb)

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

func TestPersistListingObservationStoresDetailTimestampOnFirstObservation(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	observedAt := time.Now().UTC()
	detailAt := observedAt.Add(-time.Minute)
	description := "Fetched once"
	availability := "available"
	listing := DiscoveredListing{
		Source:       "cashConverters",
		CanonicalUrl: "https://example.test/item/1",
		Title:        "Item",
		Description:  &description,
		Availability: &availability,
		LastDetailAt: &detailAt,
	}

	listingID, err := dbClient.PersistListingObservation(context.Background(), listing, observedAt)
	if err != nil {
		t.Fatalf("persist listing observation: %v", err)
	}

	persisted, err := dbClient.Listing.WithContext(context.Background()).Where(dbClient.Listing.ID.Eq(listingID)).First()
	if err != nil {
		t.Fatalf("read persisted listing: %v", err)
	}
	if persisted.LastDetailAt == nil || !persisted.LastDetailAt.Equal(detailAt) {
		t.Fatalf("last detail time = %v; expected %v", persisted.LastDetailAt, detailAt)
	}
}

func TestPersistListingObservationSkipsFreshListingUpdateButKeepsObservations(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	observedAt := time.Now().UTC()
	availability := "available"
	firstPrice := 100.0
	listing := DiscoveredListing{
		Source:       "ebay",
		CanonicalUrl: "https://example.test/item/1",
		Title:        "Item",
		Price:        &firstPrice,
		TotalPrice:   &firstPrice,
		Availability: &availability,
	}

	listingID, err := dbClient.PersistListingObservation(ctx, listing, observedAt)
	if err != nil {
		t.Fatalf("persist first observation: %v", err)
	}

	secondPrice := 90.0
	listing.Price = &secondPrice
	listing.TotalPrice = &secondPrice
	if _, err := dbClient.PersistListingObservation(ctx, listing, observedAt.Add(time.Minute)); err != nil {
		t.Fatalf("persist second observation: %v", err)
	}

	persisted, err := dbClient.Listing.WithContext(ctx).Where(dbClient.Listing.ID.Eq(listingID)).First()
	if err != nil {
		t.Fatalf("read persisted listing: %v", err)
	}
	if !persisted.LastSeenAt.Equal(observedAt) {
		t.Fatalf("last seen = %v; expected unchanged %v", persisted.LastSeenAt, observedAt)
	}

	var observationCount int64
	if err := dbClient.UnderlyingDB().WithContext(ctx).
		Model(&models.ListingObservation{}).
		Where("listingId = ?", listingID).
		Count(&observationCount).Error; err != nil {
		t.Fatalf("count observations: %v", err)
	}
	if observationCount != 2 {
		t.Fatalf("observation count = %d; expected 2", observationCount)
	}
}

func TestPersistListingObservationSkipsFreshUnchangedObservation(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	observedAt := time.Now().UTC()
	availability := "available"
	price := 100.0
	listing := DiscoveredListing{
		Source:       "ebay",
		CanonicalUrl: "https://example.test/item/1",
		Title:        "Item",
		Price:        &price,
		TotalPrice:   &price,
		Availability: &availability,
	}

	listingID, err := dbClient.PersistListingObservation(ctx, listing, observedAt)
	if err != nil {
		t.Fatalf("persist first observation: %v", err)
	}
	if _, err := dbClient.PersistListingObservation(ctx, listing, observedAt.Add(time.Minute)); err != nil {
		t.Fatalf("persist second observation: %v", err)
	}

	var observationCount int64
	if err := dbClient.UnderlyingDB().WithContext(ctx).
		Model(&models.ListingObservation{}).
		Where("listingId = ?", listingID).
		Count(&observationCount).Error; err != nil {
		t.Fatalf("count observations: %v", err)
	}
	if observationCount != 1 {
		t.Fatalf("observation count = %d; expected 1", observationCount)
	}
}

func TestPersistListingBatchSkipsFreshUnchangedObservationAndState(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	observedAt := time.Now().UTC()
	availability := "available"
	price := 100.0
	found := DiscoveredListing{
		Source:       "cashConverters",
		CanonicalUrl: "https://example.test/item/1",
		Title:        "Item",
		Price:        &price,
		TotalPrice:   &price,
		Availability: &availability,
	}
	listingID := StableListingId(found.Source, found.CanonicalUrl)

	queryObj := models.SearchQuery{ID: "batch-query", CreatedAt: observedAt}
	if err := dbClient.SearchQuery.WithContext(ctx).Create(&queryObj); err != nil {
		t.Fatalf("create query: %v", err)
	}

	initialState := &models.QueryListingState{
		QueryId:             queryObj.ID,
		ListingId:           listingID,
		Source:              found.Source,
		Status:              models.QueryListingStateStatusRejected,
		LastEvaluatedAt:     observedAt,
		LastRejectedReason:  stringPtr("AboveMaxPrice"),
		LowestObservedPrice: &price,
	}
	if err := dbClient.PersistListingBatch(ctx, []DiscoveredListing{found}, map[string]*models.Listing{}, map[string]*models.ListingObservation{}, map[string]*models.QueryListingState{}, []*models.QueryListingState{initialState}, observedAt); err != nil {
		t.Fatalf("persist initial batch: %v", err)
	}

	existingListings, err := dbClient.LoadListings(ctx, []string{listingID})
	if err != nil {
		t.Fatalf("load listings: %v", err)
	}
	latestObservations, err := dbClient.LoadLatestListingObservations(ctx, []string{listingID})
	if err != nil {
		t.Fatalf("load latest observations: %v", err)
	}
	existingStates, err := dbClient.LoadQueryListingStates(ctx, queryObj.ID, []string{listingID})
	if err != nil {
		t.Fatalf("load states: %v", err)
	}

	nextState := &models.QueryListingState{
		QueryId:             queryObj.ID,
		ListingId:           listingID,
		Source:              found.Source,
		Status:              models.QueryListingStateStatusRejected,
		LastEvaluatedAt:     observedAt.Add(time.Minute),
		LastRejectedReason:  stringPtr("AboveMaxPrice"),
		LowestObservedPrice: &price,
	}
	if err := dbClient.PersistListingBatch(ctx, []DiscoveredListing{found}, existingListings, latestObservations, existingStates, []*models.QueryListingState{nextState}, observedAt.Add(time.Minute)); err != nil {
		t.Fatalf("persist unchanged batch: %v", err)
	}

	var observationCount int64
	if err := dbClient.UnderlyingDB().WithContext(ctx).
		Model(&models.ListingObservation{}).
		Where("listingId = ?", listingID).
		Count(&observationCount).Error; err != nil {
		t.Fatalf("count observations: %v", err)
	}
	if observationCount != 1 {
		t.Fatalf("observation count = %d; expected 1", observationCount)
	}

	state, err := dbClient.GetQueryListingState(ctx, queryObj.ID, listingID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if !state.LastEvaluatedAt.Equal(observedAt) {
		t.Fatalf("last evaluated = %v; expected unchanged %v", state.LastEvaluatedAt, observedAt)
	}
}

func TestPersistListingBatchKeepsFreshPriceChangeObservationAndState(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	observedAt := time.Now().UTC()
	availability := "available"
	firstPrice := 100.0
	found := DiscoveredListing{
		Source:       "cashConverters",
		CanonicalUrl: "https://example.test/item/1",
		Title:        "Item",
		Price:        &firstPrice,
		TotalPrice:   &firstPrice,
		Availability: &availability,
	}
	listingID := StableListingId(found.Source, found.CanonicalUrl)

	queryObj := models.SearchQuery{ID: "batch-query", CreatedAt: observedAt}
	if err := dbClient.SearchQuery.WithContext(ctx).Create(&queryObj); err != nil {
		t.Fatalf("create query: %v", err)
	}

	initialState := &models.QueryListingState{
		QueryId:             queryObj.ID,
		ListingId:           listingID,
		Source:              found.Source,
		Status:              models.QueryListingStateStatusMatched,
		LastEvaluatedAt:     observedAt,
		LastMatchedAt:       &observedAt,
		FirstMatchedAt:      &observedAt,
		LowestObservedPrice: &firstPrice,
	}
	if err := dbClient.PersistListingBatch(ctx, []DiscoveredListing{found}, map[string]*models.Listing{}, map[string]*models.ListingObservation{}, map[string]*models.QueryListingState{}, []*models.QueryListingState{initialState}, observedAt); err != nil {
		t.Fatalf("persist initial batch: %v", err)
	}

	existingListings, err := dbClient.LoadListings(ctx, []string{listingID})
	if err != nil {
		t.Fatalf("load listings: %v", err)
	}
	latestObservations, err := dbClient.LoadLatestListingObservations(ctx, []string{listingID})
	if err != nil {
		t.Fatalf("load latest observations: %v", err)
	}
	existingStates, err := dbClient.LoadQueryListingStates(ctx, queryObj.ID, []string{listingID})
	if err != nil {
		t.Fatalf("load states: %v", err)
	}

	secondPrice := 90.0
	found.Price = &secondPrice
	found.TotalPrice = &secondPrice
	evaluatedAt := observedAt.Add(time.Minute)
	nextState := &models.QueryListingState{
		QueryId:             queryObj.ID,
		ListingId:           listingID,
		Source:              found.Source,
		Status:              models.QueryListingStateStatusMatched,
		LastEvaluatedAt:     evaluatedAt,
		LastMatchedAt:       &evaluatedAt,
		FirstMatchedAt:      initialState.FirstMatchedAt,
		LowestObservedPrice: &secondPrice,
	}
	if err := dbClient.PersistListingBatch(ctx, []DiscoveredListing{found}, existingListings, latestObservations, existingStates, []*models.QueryListingState{nextState}, evaluatedAt); err != nil {
		t.Fatalf("persist changed batch: %v", err)
	}

	var observationCount int64
	if err := dbClient.UnderlyingDB().WithContext(ctx).
		Model(&models.ListingObservation{}).
		Where("listingId = ?", listingID).
		Count(&observationCount).Error; err != nil {
		t.Fatalf("count observations: %v", err)
	}
	if observationCount != 2 {
		t.Fatalf("observation count = %d; expected 2", observationCount)
	}

	state, err := dbClient.GetQueryListingState(ctx, queryObj.ID, listingID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.LowestObservedPrice == nil || *state.LowestObservedPrice != secondPrice {
		t.Fatalf("lowest observed price = %v; expected %v", state.LowestObservedPrice, secondPrice)
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

	queryObj := models.SearchQuery{ID: "test-query", CreatedAt: time.Now()}
	if err := dbClient.SearchQuery.WithContext(ctx).Create(&queryObj); err != nil {
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

	dec1, err := dbClient.EvaluateListingForQuery(ctx, queryObj.ID, listingId, "ebay", firstListing.TotalPrice, match1, time.Now(), 0)
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

	dec2, err := dbClient.EvaluateListingForQuery(ctx, queryObj.ID, listingId, "ebay", secondListing.TotalPrice, match2, time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}
	if dec2.Type != DecisionTypeNotify || dec2.Reason != DecisionReasonPriceDroppedIntoRange {
		t.Errorf("Expected Notify PriceDroppedIntoRange, got type %q, reason %q", dec2.Type, dec2.Reason)
	}

	dec3, err := dbClient.EvaluateListingForQuery(ctx, queryObj.ID, listingId, "ebay", secondListing.TotalPrice, match2, time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}
	if dec3.Type != DecisionTypeDoNotNotify || dec3.Reason != DecisionReasonAlreadyNotified {
		t.Errorf("Expected DoNotNotify AlreadyNotified, got type %q, reason %q", dec3.Type, dec3.Reason)
	}

	count, err := dbClient.ListingObservation.WithContext(ctx).Where(dbClient.ListingObservation.ListingId.Eq(listingId)).Count()
	if err != nil {
		t.Fatalf("Failed to count listing observations: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 observations, got %d", count)
	}
}

func TestEvaluateListingForQuerySkipsUnchangedFreshStateWrite(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	queryObj := models.SearchQuery{ID: "test-query-fresh-state", CreatedAt: time.Now()}
	if err := dbClient.SearchQuery.WithContext(ctx).Create(&queryObj); err != nil {
		t.Fatalf("create query: %v", err)
	}

	listing := DiscoveredListing{
		Source:       "ebay",
		CanonicalUrl: "https://example.test/fresh-state",
		Title:        "Fresh state",
	}
	listingID, err := dbClient.PersistListingObservation(ctx, listing, time.Now())
	if err != nil {
		t.Fatalf("persist listing: %v", err)
	}

	firstEvaluatedAt := time.Now().UTC()
	decision, err := dbClient.EvaluateListingForQuery(
		ctx,
		queryObj.ID,
		listingID,
		"ebay",
		nil,
		MatchResult{Type: MatchTypeRejected, Reason: "AboveMaxPrice"},
		firstEvaluatedAt,
		0,
	)
	if err != nil {
		t.Fatalf("first evaluation: %v", err)
	}
	if decision.Type != DecisionTypeDoNotNotify {
		t.Fatalf("first decision = %v; expected do not notify", decision.Type)
	}

	_, err = dbClient.EvaluateListingForQuery(
		ctx,
		queryObj.ID,
		listingID,
		"ebay",
		nil,
		MatchResult{Type: MatchTypeRejected, Reason: "AboveMaxPrice"},
		firstEvaluatedAt.Add(time.Minute),
		0,
	)
	if err != nil {
		t.Fatalf("second evaluation: %v", err)
	}

	state, err := dbClient.GetQueryListingState(ctx, queryObj.ID, listingID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state == nil {
		t.Fatal("state was not persisted")
	}
	if !state.LastEvaluatedAt.Equal(firstEvaluatedAt) {
		t.Fatalf("last evaluated = %v; expected unchanged %v", state.LastEvaluatedAt, firstEvaluatedAt)
	}
}

func TestFurtherPriceDrops(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	queryObj := models.SearchQuery{ID: "test-query-drop", CreatedAt: time.Now()}
	if err := dbClient.SearchQuery.WithContext(ctx).Create(&queryObj); err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

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

	dec1, err := dbClient.EvaluateListingForQuery(ctx, queryObj.ID, listingId, "ebay", firstListing.TotalPrice, match, time.Now(), 0.05)
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
	dec2, err := dbClient.EvaluateListingForQuery(ctx, queryObj.ID, listingId, "ebay", secondListing.TotalPrice, match, time.Now(), 0.05)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}
	if dec2.Type != DecisionTypeNotify || dec2.Reason != DecisionReasonFurtherPriceDrop {
		t.Errorf("Expected Notify FurtherPriceDrop, got type %q, reason %q", dec2.Type, dec2.Reason)
	}
}

func TestEvaluateListingForQueryPreservesNotificationMetadataOnRejection(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	queryObj := models.SearchQuery{ID: "test-query-preserve", CreatedAt: time.Now()}
	if err := dbClient.SearchQuery.WithContext(ctx).Create(&queryObj); err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	listingId := StableListingId("cashConverters", "https://example.test/preserve")
	firstPrice := 80.0
	firstAt := time.Now().UTC().Add(-time.Hour)

	dec1, err := dbClient.EvaluateListingForQuery(
		ctx,
		queryObj.ID,
		listingId,
		"cashConverters",
		&firstPrice,
		MatchResult{Type: MatchTypeMatched},
		firstAt,
		0,
	)
	if err != nil {
		t.Fatalf("Failed to evaluate first match: %v", err)
	}
	if dec1.Type != DecisionTypeNotify {
		t.Fatalf("Expected first match to notify, got %q", dec1.Type)
	}

	rejectedPrice := 140.0
	secondAt := firstAt.Add(time.Hour)
	dec2, err := dbClient.EvaluateListingForQuery(
		ctx,
		queryObj.ID,
		listingId,
		"cashConverters",
		&rejectedPrice,
		MatchResult{Type: MatchTypeRejected, Reason: "AboveMaxPrice"},
		secondAt,
		0,
	)
	if err != nil {
		t.Fatalf("Failed to evaluate rejection: %v", err)
	}
	if dec2.Type != DecisionTypeDoNotNotify || dec2.Reason != DecisionReasonRejected {
		t.Fatalf("Expected rejection without notification, got type %q reason %q", dec2.Type, dec2.Reason)
	}

	state, err := dbClient.GetQueryListingState(ctx, queryObj.ID, listingId)
	if err != nil {
		t.Fatalf("Failed to read listing state: %v", err)
	}
	if state == nil {
		t.Fatal("Expected persisted listing state")
	}
	if state.Status != ListingStateStatusRejected {
		t.Fatalf("Expected rejected status, got %q", state.Status)
	}
	if state.FirstMatchedAt == nil || !state.FirstMatchedAt.Equal(firstAt) {
		t.Fatalf("first matched time = %v; expected %v", state.FirstMatchedAt, firstAt)
	}
	if state.LastMatchedAt == nil || !state.LastMatchedAt.Equal(firstAt) {
		t.Fatalf("last matched time = %v; expected %v", state.LastMatchedAt, firstAt)
	}
	if state.LastNotifiedAt == nil || !state.LastNotifiedAt.Equal(firstAt) {
		t.Fatalf("last notified time = %v; expected %v", state.LastNotifiedAt, firstAt)
	}
	if state.LastNotifiedTotalPrice == nil || *state.LastNotifiedTotalPrice != firstPrice {
		t.Fatalf("last notified total price = %v; expected %v", state.LastNotifiedTotalPrice, firstPrice)
	}
}

func TestEvaluateListingForQueryNotifiesWhenPreviouslyRejectedThenMatched(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	queryObj := models.SearchQuery{ID: "test-query-rejected-then-matched", CreatedAt: time.Now()}
	if err := dbClient.SearchQuery.WithContext(ctx).Create(&queryObj); err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	listingId := StableListingId("cashConverters", "https://example.test/rejected-then-matched")
	rejectedPrice := 140.0
	firstAt := time.Now().UTC().Add(-time.Hour)

	_, err := dbClient.EvaluateListingForQuery(
		ctx,
		queryObj.ID,
		listingId,
		"cashConverters",
		&rejectedPrice,
		MatchResult{Type: MatchTypeRejected, Reason: "AboveMaxPrice"},
		firstAt,
		0,
	)
	if err != nil {
		t.Fatalf("Failed to evaluate rejection: %v", err)
	}

	matchedPrice := 80.0
	secondAt := firstAt.Add(time.Hour)
	decision, err := dbClient.EvaluateListingForQuery(
		ctx,
		queryObj.ID,
		listingId,
		"cashConverters",
		&matchedPrice,
		MatchResult{Type: MatchTypeMatched},
		secondAt,
		0,
	)
	if err != nil {
		t.Fatalf("Failed to evaluate match: %v", err)
	}
	if decision.Type != DecisionTypeNotify || decision.Reason != DecisionReasonPriceDroppedIntoRange {
		t.Fatalf("Expected notification after rejected listing matched, got type %q reason %q", decision.Type, decision.Reason)
	}

	state, err := dbClient.GetQueryListingState(ctx, queryObj.ID, listingId)
	if err != nil {
		t.Fatalf("Failed to read listing state: %v", err)
	}
	if state == nil {
		t.Fatal("Expected persisted listing state")
	}
	if state.FirstMatchedAt == nil || !state.FirstMatchedAt.Equal(secondAt) {
		t.Fatalf("first matched time = %v; expected %v", state.FirstMatchedAt, secondAt)
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}
