package scanners

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBuildCcNewestPageURL(t *testing.T) {
	urlStr := buildCcCatalogPageURL(ccCatalogSortNewest, 2)
	expected := "https://www.cashconverters.com.au/c3api/search/results?Sort=newest&page=2&query="
	if urlStr != expected {
		t.Errorf("buildCcCatalogPageURL(newest, 2) = %s; expected %s", urlStr, expected)
	}
}

func TestPlanDiscoveryPagesContinuesCatalogSweep(t *testing.T) {
	scanner := &CashConvertersScanner{}
	state := &models.CashConvertersScanState{TailNextPage: 412}

	pages := scanner.planDiscoveryPages(state)
	if len(pages) != ccTailPagesPerPass {
		t.Fatalf("planned %d pages; expected %d", len(pages), ccTailPagesPerPass)
	}

	firstCatalog := pages[0]
	if firstCatalog.sort != ccCatalogSortPrice || firstCatalog.page != 412 {
		t.Fatalf("first catalog page was planned as %+v", firstCatalog)
	}
}

func TestNormalizeCcCanonicalURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/shop/item/023500607432",
			expected: "https://www.cashconverters.com.au/shop/item/023500607432",
		},
		{
			input:    "https://www.cashconverters.com.au/shop/item/023500607432?utm_source=google&ref=123",
			expected: "https://www.cashconverters.com.au/shop/item/023500607432",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			res := normalizeCcCanonicalURL(tc.input)
			if res != tc.expected {
				t.Errorf("normalizeCcCanonicalURL(%q) = %s; expected %s", tc.input, res, tc.expected)
			}
		})
	}
}

func TestExtractCcExternalID(t *testing.T) {
	item := CcApiProductItem{Url: "/shop/item/023500607432"}
	canonical := "https://www.cashconverters.com.au/shop/item/023500607432"

	idPtr := extractCcExternalID(item, canonical)
	if idPtr == nil || *idPtr != "023500607432" {
		t.Fatalf("expected external ID '023500607432', got %v", idPtr)
	}
}

func TestCashConvertersFilterMatcher(t *testing.T) {
	matcher := CashConvertersFilterMatcher{
		Required:     []string{"snapdragon", "laptop"},
		RequiredMode: "all_words",
		Excluded:     []string{"broken", "damaged"},
		ExcludedMode: "any_words",
		MinPrice:     floatPtr(100),
		MaxPrice:     floatPtr(1000),
	}

	// 1. Matches cleanly
	res1 := matcher.Match("Snapdragon X Laptop", "Awesome 16GB RAM model in great condition", 500)
	if !res1.Matched {
		t.Errorf("expected clean match, got reject reason: %v", res1.RejectReason)
	}

	// 2. Fails missing phrase
	res2 := matcher.Match("Intel Core i7 Laptop", "Fast laptop", 500)
	if res2.Matched || res2.RejectReason == nil || *res2.RejectReason != "MissingRequiredPhrase" {
		t.Errorf("expected MissingRequiredPhrase, got %v", res2)
	}

	// 3. Fails excluded phrase
	res3 := matcher.Match("Snapdragon X Laptop", "Broken screen, for parts", 500)
	if res3.Matched || res3.RejectReason == nil || *res3.RejectReason != "ExcludedPhrase" {
		t.Errorf("expected ExcludedPhrase, got %v", res3)
	}

	// 4. Fails price range
	res4 := matcher.Match("Snapdragon X Laptop", "Good condition", 50)
	if res4.Matched || res4.RejectReason == nil || *res4.RejectReason != "BelowMinPrice" {
		t.Errorf("expected BelowMinPrice, got %v", res4)
	}
}

func TestCashConvertersPhraseMatchingCanUseWordsOrLetterSequences(t *testing.T) {
	searchable := "programming laptop with 16gb ram"

	if phrasesMatch(searchable, []string{"ram"}, "all_words") == false {
		t.Fatal("expected standalone word ram to match")
	}
	if phrasesMatch(searchable, []string{"gram"}, "all_words") {
		t.Fatal("did not expect gram inside programming to match as a whole word")
	}
	if !phrasesMatch(searchable, []string{"gram"}, "all_substring") {
		t.Fatal("expected gram to match as a letter sequence")
	}
	if !phrasesMatch("geforce rtx 4060-ti", []string{"rtx 4060"}, "all_words") {
		t.Fatal("expected a whole multi-word phrase followed by punctuation to match")
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestParseCcPrice(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"$123.45", 123.45},
		{"Price: $10.00", 10.00},
		{"15.5", 15.5},
		{"invalid", 0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			val := parseCcPrice(tc.input)
			diff := val - tc.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("parseCcPrice(%q) = %f; expected %f", tc.input, val, tc.expected)
			}
		})
	}
}

func TestParseCcShipping(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"FREE", 0},
		{"Free postage", 0},
		{"$5.00", 5.00},
		{"invalid", 0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			val := parseCcShipping(tc.input)
			diff := val - tc.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("parseCcShipping(%q) = %f; expected %f", tc.input, val, tc.expected)
			}
		})
	}
}

func TestParsePhrases(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{" xbox, Series X ,, Controller ", []string{"xbox", "series x", "controller"}},
		{"a, b, c", []string{"a", "b", "c"}},
		{"", nil},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			res := parsePhrases(tc.input)
			if len(res) == 0 && len(tc.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("parsePhrases(%q) = %v; expected %v", tc.input, res, tc.expected)
			}
		})
	}
}

func TestCashConvertersFullScannerLoopWithMockHTTP(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open memory db: %v", err)
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
		t.Fatalf("auto migrate: %v", err)
	}

	dbClient := qry.Use(gdb)

	if err := gdb.Create(&models.Globals{ID: "1", CashConverters: true}).Error; err != nil {
		t.Fatalf("create globals: %v", err)
	}

	ctx := context.Background()
	_, err = dbClient.CreateCashConvertersQuery(ctx, false, qry.CashConvertersQueryInput{
		Url:             "snapdragon",
		RequiredPhrases: "snapdragon",
		MaxPrice:        floatPtr(800),
	})
	if err != nil {
		t.Fatalf("create query: %v", err)
	}

	// Mock server handling search results API and detail page fetches
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/c3api/search/results" {
			page := r.URL.Query().Get("page")
			w.Header().Set("Content-Type", "application/json")
			if page == "1" {
				fmt.Fprint(w, `{
					"WasSuccessful": true,
					"Value": {
						"ProductList": {
							"ProductListItems": [
								{
									"Title": "Snapdragon Laptop",
									"Sp": "$500.00",
									"ShippingCost": "$10.00",
									"AbsoluteImageUrl": "https://example.com/item1.jpg",
									"Url": "/shop/item/023500607432"
								}
							]
						}
					}
				}`)
			} else {
				fmt.Fprint(w, `{
					"WasSuccessful": true,
					"Value": {
						"ProductList": {
							"ProductListItems": []
						}
					}
				}`)
			}
			return
		}

		if r.URL.Path == "/shop/item/023500607432" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html>
				<head>
					<title>Snapdragon Laptop</title>
					<meta name="description" content="Features Qualcomm Snapdragon X Plus 16GB RAM">
				</head>
				<body>
					<h1 class="product-title">Snapdragon Laptop</h1>
				</body>
			</html>`)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	scanner := NewCashConvertersScanner(dbClient)
	scanner.baseURL = server.URL

	// Run scan pass 1
	notifs, err := scanner.Scan(ctx)
	if err != nil {
		t.Fatalf("scanner.Scan failed: %v", err)
	}

	// Verify listing persisted
	var lst models.Listing
	listingID := qry.StableListingId("cashConverters", "https://www.cashconverters.com.au/shop/item/023500607432")
	if err := gdb.Where("id = ?", listingID).First(&lst).Error; err != nil {
		t.Fatalf("expected listing %s to be persisted, error: %v", listingID, err)
	}

	if lst.Title != "Snapdragon Laptop" {
		t.Errorf("expected title 'Snapdragon Laptop', got %s", lst.Title)
	}

	// Verify search doc updated with description
	var searchDoc models.CashConvertersSearchDoc
	if err := gdb.Where("listingId = ?", listingID).First(&searchDoc).Error; err != nil {
		t.Fatalf("expected search doc for %s, error: %v", listingID, err)
	}

	if searchDoc.Description == "" {
		t.Errorf("expected description to be enriched from detail job")
	}

	// Verify notifications generated
	if len(notifs) != 1 {
		t.Fatalf("expected 1 notification generated for snapdragon laptop, got %d", len(notifs))
	}
}

func TestCashConvertersScannerAPIErrorResilience(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open memory db: %v", err)
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
	_ = gdb.Create(&models.Globals{ID: "1", CashConverters: true})

	// Server returning 500 internal server error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	scanner := NewCashConvertersScanner(dbClient)
	scanner.baseURL = server.URL

	ctx := context.Background()
	_, err = scanner.Scan(ctx)
	if err != nil {
		t.Fatalf("expected scanner to handle HTTP 500 error gracefully without crashing, got: %v", err)
	}

	state, err := dbClient.GetCashConvertersScanState(ctx)
	if err != nil {
		t.Fatalf("get scan state: %v", err)
	}

	if state.LastError == nil {
		t.Errorf("expected state.LastError to record API error")
	}
}
