package query

import (
	"context"
	"testing"
	"time"

	"dealscanner/internal/models"
)

func TestGetLastFoundItemsBatch(t *testing.T) {
	dbClient, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Create mock listings
	listing1 := models.Listing{
		ID:           "list1",
		Source:       "ebay",
		ExternalId:   stringPtr("ext1"),
		CanonicalUrl: "https://ebay.com.au/1",
		Title:        "Xbox Controller Blue",
	}
	listing2 := models.Listing{
		ID:           "list2",
		Source:       "ebay",
		ExternalId:   stringPtr("ext2"),
		CanonicalUrl: "https://ebay.com.au/2",
		Title:        "Xbox Controller Black",
	}
	listing3 := models.Listing{
		ID:           "list3",
		Source:       "ebay",
		ExternalId:   stringPtr("ext3"),
		CanonicalUrl: "https://ebay.com.au/3",
		Title:        "Xbox Controller Red",
	}

	err := dbClient.db.Create(&listing1).Error
	if err != nil {
		t.Fatalf("Failed to create listing1: %v", err)
	}
	dbClient.db.Create(&listing2)
	dbClient.db.Create(&listing3)

	// 2. Create mock QueryListingStates
	now := time.Now().Truncate(time.Second)
	state1 := models.QueryListingState{
		QueryId:                "queryA",
		ListingId:              "list1",
		Source:                 "ebay",
		Status:                 "notified",
		LastMatchedAt:          &now,
		LastNotifiedTotalPrice: floatPtr(50.0),
	}

	t2 := now.Add(-10 * time.Minute)
	state2 := models.QueryListingState{
		QueryId:                "queryA",
		ListingId:              "list2",
		Source:                 "ebay",
		Status:                 "matched",
		LastMatchedAt:          &t2,
		LastNotifiedTotalPrice: floatPtr(45.0),
	}

	t3 := now.Add(10 * time.Minute)
	state3 := models.QueryListingState{
		QueryId:                "queryA",
		ListingId:              "list3",
		Source:                 "ebay",
		Status:                 "notified",
		LastMatchedAt:          &t3,
		LastNotifiedTotalPrice: floatPtr(55.0),
	}

	// State for another query
	stateB := models.QueryListingState{
		QueryId:                "queryB",
		ListingId:              "list1",
		Source:                 "ebay",
		Status:                 "notified",
		LastMatchedAt:          &now,
		LastNotifiedTotalPrice: floatPtr(50.0),
	}

	dbClient.db.Create(&state1)
	dbClient.db.Create(&state2)
	dbClient.db.Create(&state3)
	dbClient.db.Create(&stateB)

	// 3. Test GetLastFoundItemsBatch
	queryIds := []string{"queryA", "queryB"}
	res, err := dbClient.GetLastFoundItemsBatch(ctx, queryIds, 2)
	if err != nil {
		t.Fatalf("GetLastFoundItemsBatch failed: %v", err)
	}

	// queryA should have 2 items (limit is 2), ordered by LastMatchedAt DESC (state3 then state1)
	itemsA := res["queryA"]
	if len(itemsA) != 2 {
		t.Fatalf("Expected 2 items for queryA, got %d", len(itemsA))
	}

	if itemsA[0].Title != "Xbox Controller Red" {
		t.Errorf("Expected most recent item to be Red, got %s", itemsA[0].Title)
	}
	if itemsA[0].Price != 55.0 {
		t.Errorf("Expected Red price to be 55.0, got %f", itemsA[0].Price)
	}

	if itemsA[1].Title != "Xbox Controller Blue" {
		t.Errorf("Expected second most recent item to be Blue, got %s", itemsA[1].Title)
	}

	// queryB should have 1 item
	itemsB := res["queryB"]
	if len(itemsB) != 1 {
		t.Fatalf("Expected 1 item for queryB, got %d", len(itemsB))
	}
	if itemsB[0].Title != "Xbox Controller Blue" {
		t.Errorf("Expected queryB item to be Blue, got %s", itemsB[0].Title)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
