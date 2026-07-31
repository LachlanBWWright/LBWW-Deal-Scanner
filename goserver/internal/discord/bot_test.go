package discord

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	qry "dealscanner/internal/db/query"
)

func TestTruncateDiscordMessageLeavesValidContentUnchanged(t *testing.T) {
	content := "Saved queries"

	if got := truncateDiscordMessage(content); got != content {
		t.Fatalf("truncateDiscordMessage() = %q, want %q", got, content)
	}
}

func TestTruncateStatusLeavesValidContentUnchanged(t *testing.T) {
	status := "Scanning Ebay | Cash Converters (00:20): NOT FOUND"
	if actual := truncateStatus(status); actual != status {
		t.Fatalf("truncateStatus() = %q, want %q", actual, status)
	}
}

func TestTruncateStatusRespectsDiscordCharacterLimit(t *testing.T) {
	status := strings.Repeat("🔎", discordStatusCharacterLimit+10)
	actual := truncateStatus(status)

	if count := utf8.RuneCountInString(actual); count != discordStatusCharacterLimit {
		t.Fatalf(
			"truncateStatus() returned %d characters, want %d",
			count,
			discordStatusCharacterLimit,
		)
	}
	if !strings.HasSuffix(actual, "…") {
		t.Fatalf("truncateStatus() = %q, want ellipsis suffix", actual)
	}
}

func TestTruncateDiscordMessageFitsDiscordLimit(t *testing.T) {
	content := strings.Repeat("query 🔎 ", 400)

	got := truncateDiscordMessage(content)

	if !utf8.ValidString(got) {
		t.Fatal("truncateDiscordMessage() returned invalid UTF-8")
	}
	if runeCount := utf8.RuneCountInString(got); runeCount > discordMessageLimit {
		t.Fatalf("truncateDiscordMessage() returned %d characters, limit is %d", runeCount, discordMessageLimit)
	}
	if !strings.HasSuffix(got, "… More query details were omitted from this page.") {
		t.Fatalf("truncateDiscordMessage() did not include the truncation notice: %q", got)
	}
}

func TestPaginateQueryEntriesFitsAsManyCompleteEntriesAsPossible(t *testing.T) {
	entries := []string{
		strings.Repeat("a", 900),
		strings.Repeat("b", 900),
		strings.Repeat("c", 900),
	}

	pages := paginateQueryEntries("Saved Queries", entries)

	if len(pages) != 2 {
		t.Fatalf("paginateQueryEntries() returned %d pages, want 2", len(pages))
	}
	if !strings.Contains(pages[0], entries[0]+entries[1]) {
		t.Fatal("first page did not contain both entries that fit")
	}
	if !strings.Contains(pages[1], entries[2]) {
		t.Fatal("second page did not contain the remaining entry")
	}
	for _, page := range pages {
		if utf8.RuneCountInString(page) > discordMessageLimit {
			t.Fatalf("page exceeded Discord limit: %d", utf8.RuneCountInString(page))
		}
	}
}

func TestPaginateQueryEntriesAlwaysReturnsAPage(t *testing.T) {
	pages := paginateQueryEntries("Saved Queries", nil)

	if len(pages) != 1 {
		t.Fatalf("paginateQueryEntries() returned %d pages, want 1", len(pages))
	}
}

func TestPaginateQueryEntriesTruncatesOnlyOversizedSingleEntry(t *testing.T) {
	pages := paginateQueryEntries("Saved Queries", []string{strings.Repeat("🔎", discordMessageLimit)})

	if len(pages) != 1 {
		t.Fatalf("paginateQueryEntries() returned %d pages, want 1", len(pages))
	}
	if utf8.RuneCountInString(pages[0]) > discordMessageLimit {
		t.Fatalf("page exceeded Discord limit: %d", utf8.RuneCountInString(pages[0]))
	}
	if !strings.Contains(pages[0], "More query details were omitted") {
		t.Fatal("oversized entry did not include truncation notice")
	}
}

func TestCashSearchInputRoundTripPreservesPaginationQuery(t *testing.T) {
	minPrice := 10.5
	maxPrice := 250.0
	input := qry.CashConvertersCatalogSearchInput{
		Text:          "mario kart",
		MinPrice:      &minPrice,
		MaxPrice:      &maxPrice,
		AvailableOnly: true,
		Limit:         25,
		Required:      "console, boxed",
		RequiredMode:  "all_words",
		Excluded:      "damaged, parts",
		ExcludedMode:  "any_substring",
		Sort:          "price_asc",
	}

	encoded, err := encodeCashSearchInput(input)
	if err != nil {
		t.Fatalf("encodeCashSearchInput() error = %v", err)
	}
	decoded, err := decodeCashSearchInput(encoded)
	if err != nil {
		t.Fatalf("decodeCashSearchInput() error = %v", err)
	}

	if decoded.Text != input.Text ||
		decoded.MinPrice == nil || *decoded.MinPrice != minPrice ||
		decoded.MaxPrice == nil || *decoded.MaxPrice != maxPrice ||
		decoded.AvailableOnly != input.AvailableOnly ||
		decoded.Limit != input.Limit ||
		decoded.Required != input.Required ||
		decoded.RequiredMode != input.RequiredMode ||
		decoded.Excluded != input.Excluded ||
		decoded.ExcludedMode != input.ExcludedMode ||
		decoded.Sort != input.Sort {
		t.Fatalf("decoded input = %#v, want %#v", decoded, input)
	}
}

func TestCashSearchResultsPaginateWithoutDroppingEntries(t *testing.T) {
	results := make([]qry.CashConvertersCatalogSearchResult, 25)
	for idx := range results {
		results[idx] = qry.CashConvertersCatalogSearchResult{
			Title:        "Mario item " + strings.Repeat("x", 120),
			Description:  "Complete product description " + strings.Repeat("y", 220),
			CanonicalURL: "https://www.cashconverters.com.au/shop/item/unique-" + string(rune('A'+idx)),
			TotalPrice:   float64Ptr(float64(idx + 1)),
			Availability: "available",
			SourceStatus: "available",
			LastSeenAt:   time.Date(2026, 7, 27, 1, 2, 3, 0, time.UTC),
		}
	}

	entries := formatCashSearchEntries(results)
	pages := paginateQueryEntries("Cash Converters search", entries)

	if len(pages) <= 1 {
		t.Fatalf("Cash search produced %d page, want multiple pages", len(pages))
	}
	combined := strings.Join(pages, "")
	for idx, entry := range entries {
		if !strings.Contains(combined, entry) {
			t.Fatalf("Cash search entry %d was dropped from pagination", idx+1)
		}
	}
	for _, page := range pages {
		if utf8.RuneCountInString(page) > discordMessageLimit {
			t.Fatalf("Cash search page exceeded Discord limit: %d", utf8.RuneCountInString(page))
		}
		if strings.Contains(page, "...(truncated)") {
			t.Fatal("Cash search page retained the old whole-response truncation marker")
		}
	}
}

func float64Ptr(value float64) *float64 {
	return &value
}
