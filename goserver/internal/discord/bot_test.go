package discord

import (
	"strings"
	"testing"
	"unicode/utf8"
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
