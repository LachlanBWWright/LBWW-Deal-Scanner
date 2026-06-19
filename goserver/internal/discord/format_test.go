package discord

import "testing"

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
