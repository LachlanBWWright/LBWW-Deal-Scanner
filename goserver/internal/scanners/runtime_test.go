package scanners

import (
	"context"
	"errors"
	"testing"
	"time"

	"dealscanner/internal/notifications"
)

func TestWaitForMinimumCycleDurationDoesNotWaitAfterMinimum(t *testing.T) {
	startedAt := time.Now().Add(-time.Second)
	if !waitForMinimumCycleDuration(context.Background(), startedAt, time.Second) {
		t.Fatal("expected completed minimum cycle wait")
	}
}

func TestWaitForMinimumCycleDurationStopsWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if waitForMinimumCycleDuration(ctx, time.Now(), time.Minute) {
		t.Fatal("expected cancelled cycle wait to stop")
	}
}

func TestFormatTimedStatus(t *testing.T) {
	tests := []struct {
		name     string
		scanner  string
		elapsed  time.Duration
		expected string
	}{
		{
			name:     "formats first timed update",
			scanner:  "Cash Converters",
			elapsed:  20 * time.Second,
			expected: "Scanning Cash Converters (00:20)",
		},
		{
			name:     "formats minutes and seconds",
			scanner:  "Steam Market",
			elapsed:  2*time.Minute + 7*time.Second + 900*time.Millisecond,
			expected: "Scanning the Steam Community Market (02:07)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := formatTimedStatus(test.scanner, test.elapsed)
			if actual != test.expected {
				t.Fatalf("formatTimedStatus() = %q, want %q", actual, test.expected)
			}
		})
	}
}

func TestFormatScanningStatusIncludesPreviousResult(t *testing.T) {
	previous := &scanStatusResult{
		name:    "Cash Converters",
		elapsed: 20 * time.Second,
		outcome: "NOT FOUND",
	}

	actual := formatScanningStatus("Ebay", 0, false, previous)
	expected := "Scanning Ebay | Cash Converters (00:20): NOT FOUND"
	if actual != expected {
		t.Fatalf("formatScanningStatus() = %q, want %q", actual, expected)
	}
}

func TestGetScanOutcome(t *testing.T) {
	if actual := getScanOutcome(nil, nil); actual != "NOT FOUND" {
		t.Fatalf("getScanOutcome() = %q, want NOT FOUND", actual)
	}
	if actual := getScanOutcome(
		[]notifications.AppNotification{{Kind: "deal"}},
		nil,
	); actual != "NEW ITEM FOUND" {
		t.Fatalf("getScanOutcome() = %q, want NEW ITEM FOUND", actual)
	}
	if actual := getScanOutcome(nil, errors.New("scan failed")); actual != "FAILED" {
		t.Fatalf("getScanOutcome() = %q, want FAILED", actual)
	}
}
