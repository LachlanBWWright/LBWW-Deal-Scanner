package scanners

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"dealscanner/internal/config"
	"dealscanner/internal/notifications"
)

type blockingScanner struct {
	name string
}

func (s blockingScanner) Name() string {
	return s.name
}

func (s blockingScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

type recordingProvider struct {
	received chan notifications.AppNotification
}

func (p *recordingProvider) Name() string {
	return "recording"
}

func (p *recordingProvider) IsEnabled() bool {
	return true
}

func (p *recordingProvider) Send(
	ctx context.Context,
	notification notifications.AppNotification,
) error {
	select {
	case p.received <- notification:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

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

func TestRunScannerWithTimeoutPublishesWarningToErrorNotifications(t *testing.T) {
	provider := &recordingProvider{
		received: make(chan notifications.AppNotification, 1),
	}
	runner := NewRunner(
		&config.Config{},
		nil,
		nil,
		notifications.NewNotificationService([]notifications.NotificationProvider{provider}),
		nil,
	)
	runner.scanTimeout = 10 * time.Millisecond

	_, err := runner.runScannerWithTimeout(
		context.Background(),
		blockingScanner{name: "Ebay"},
	)
	var timeoutErr scanTimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("runScannerWithTimeout() error = %v, want scanTimeoutError", err)
	}

	select {
	case notification := <-provider.received:
		if notification.Kind != "error" {
			t.Fatalf("notification kind = %q, want error", notification.Kind)
		}
		if notification.Source != "Ebay" {
			t.Fatalf("notification source = %q, want Ebay", notification.Source)
		}
		if !strings.Contains(notification.Message, "Warning:") {
			t.Fatalf("notification message = %q, want warning", notification.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for scan timeout warning")
	}
}

func TestRunScannerWithTimeoutDoesNotWarnWhenParentContextIsCancelled(t *testing.T) {
	provider := &recordingProvider{
		received: make(chan notifications.AppNotification, 1),
	}
	runner := NewRunner(
		&config.Config{},
		nil,
		nil,
		notifications.NewNotificationService([]notifications.NotificationProvider{provider}),
		nil,
	)
	runner.scanTimeout = time.Minute
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := runner.runScannerWithTimeout(ctx, blockingScanner{name: "Ebay"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runScannerWithTimeout() error = %v, want context.Canceled", err)
	}

	select {
	case notification := <-provider.received:
		t.Fatalf("unexpected notification: %#v", notification)
	case <-time.After(20 * time.Millisecond):
	}
}
