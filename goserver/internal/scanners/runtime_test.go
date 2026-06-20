package scanners

import (
	"testing"
	"time"
)

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
