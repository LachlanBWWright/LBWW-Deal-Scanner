package discord

import (
	"strings"
	"testing"

	"dealscanner/internal/db/query"
)

func TestRemoveMarkdownCodeFences(t *testing.T) {
	t.Parallel()

	got := removeMarkdownCodeFences("Before ```\ncontents\n``` after")
	want := "Before \ncontents\n after"
	if got != want {
		t.Fatalf("removeMarkdownCodeFences() = %q, want %q", got, want)
	}
}

func TestFormatQueryReferenceFormatsHTTPURLsAsDiscordLinks(t *testing.T) {
	t.Parallel()

	got := formatQueryReference("https://www.cashconverters.com.au/search-results?query=console")
	want := "<https://www.cashconverters.com.au/search-results?query=console>"
	if got != want {
		t.Fatalf("formatQueryReference() = %q, want %q", got, want)
	}
}

func TestFormatQueryReferenceFormatsNamesAsInlineCode(t *testing.T) {
	t.Parallel()

	got := formatQueryReference("AK-47 | Redline")
	want := "`AK-47 | Redline`"
	if got != want {
		t.Fatalf("formatQueryReference() = %q, want %q", got, want)
	}
}

func TestFormatQueryReferenceRemovesCharactersThatBreakDiscordFormatting(t *testing.T) {
	t.Parallel()

	got := formatQueryReference("query`\nname")
	want := "`queryˋname`"
	if got != want {
		t.Fatalf("formatQueryReference() = %q, want %q", got, want)
	}
}

func TestFormatQueriesWithFoundItemsIncludesCashConvertersFilters(t *testing.T) {
	t.Parallel()

	url := "https://www.cashconverters.com.au/search-results?query=console"
	minPrice := 100.0
	maxPrice := 450.0
	required := "ps5, controller"
	requiredMode := "all"
	excluded := "broken, parts"
	excludedMode := "any"

	got := formatCashConvertersQueryInfo(query.QueryItem{
		QueryId:           "query-1",
		Type:              "cashConverters",
		Id:                "filter-1",
		Url:               &url,
		MinPrice:          &minPrice,
		MaxPrice:          &maxPrice,
		RequiredPhrases:   &required,
		RequiredMatchMode: &requiredMode,
		ExcludePhrases:    &excluded,
		ExcludeMatchMode:  &excludedMode,
	})

	expectedParts := []string{
		"Price Range: $100.00 - $450.00",
		"Required: `ps5, controller` (all)",
		"Excluded: `broken, parts` (any)",
	}
	for _, expected := range expectedParts {
		if !strings.Contains(got, expected) {
			t.Fatalf("formatted query did not include %q: %s", expected, got)
		}
	}
}

func TestFormatQueriesWithFoundItemsShowsUnsetCashConvertersFilters(t *testing.T) {
	t.Parallel()

	url := "https://www.cashconverters.com.au/search-results?query=console"

	got := formatCashConvertersQueryInfo(query.QueryItem{
		QueryId: "query-1",
		Type:    "cashConverters",
		Id:      "filter-1",
		Url:     &url,
	})

	expectedParts := []string{
		"Price Range: Any - Any",
		"Required: `None` (all)",
		"Excluded: `None` (any)",
	}
	for _, expected := range expectedParts {
		if !strings.Contains(got, expected) {
			t.Fatalf("formatted query did not include %q: %s", expected, got)
		}
	}
}

func TestFormatQueryInfoIncludesAvailableDetailsForOtherQueryTypes(t *testing.T) {
	t.Parallel()

	url := "https://example.test/search?q=console"
	displayUrl := "https://example.test/market/item"
	name := "AK-47 | Redline"
	minPrice := 25.0
	maxPrice := 250.0
	minFloat := 0.01
	maxFloat := 0.15

	tests := []struct {
		name     string
		item     query.QueryItem
		expected []string
	}{
		{
			name: "ebay",
			item: query.QueryItem{Type: "ebay", Id: url, Url: &url, MaxPrice: &maxPrice},
			expected: []string{
				"URL: <https://example.test/search?q=console>",
				"Max Price: $250.00",
			},
		},
		{
			name: "gumtree",
			item: query.QueryItem{Type: "gumtree", Id: url, Url: &url, MaxPrice: &maxPrice},
			expected: []string{
				"URL: <https://example.test/search?q=console>",
				"Max Price: $250.00",
			},
		},
		{
			name: "salvos",
			item: query.QueryItem{Type: "salvos", Id: name, Name: &name, MinPrice: &minPrice, MaxPrice: &maxPrice},
			expected: []string{
				"Name: `AK-47 | Redline`",
				"Price Range: $25.00 - $250.00",
			},
		},
		{
			name: "csMarket",
			item: query.QueryItem{Type: "csMarket", Id: url, Url: &url, DisplayUrl: &displayUrl, MaxPrice: &maxPrice, MaxFloat: &maxFloat},
			expected: []string{
				"URL: <https://example.test/search?q=console>",
				"Display URL: <https://example.test/market/item>",
				"Max Price: $250.00",
				"Max Float: 0.15000",
			},
		},
		{
			name: "csTradeBot",
			item: query.QueryItem{Type: "csTradeBot", Id: name, Name: &name, MaxPrice: &maxPrice, MinFloat: &minFloat, MaxFloat: &maxFloat},
			expected: []string{
				"Name: `AK-47 | Redline`",
				"Max Price: $250.00",
				"Float Range: 0.01000 - 0.15000",
			},
		},
		{
			name: "steamMarket",
			item: query.QueryItem{Type: "steamMarket", Id: name, Name: &name, DisplayUrl: &displayUrl, MaxPrice: &maxPrice},
			expected: []string{
				"Name: `AK-47 | Redline`",
				"Max Price: $250.00",
				"Market Link: <https://example.test/market/item>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatQueryInfo(tt.item)
			for _, expected := range tt.expected {
				if !strings.Contains(got, expected) {
					t.Fatalf("formatted query did not include %q: %s", expected, got)
				}
			}
		})
	}
}
