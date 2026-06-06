package scanners

import (
	"testing"
)

func TestParsePrice(t *testing.T) {
	tests := []struct {
		input        string
		expectedVal  float64
		expectedCurr string
	}{
		{"AU $69.95", 69.95, "AUD"},
		{"$59.00", 91.45, "USD"}, // 59.00 * 1.55 = 91.45
		{"AU $69.95 to $80.00", 69.95, "AUD"},
		{"invalid", 0, ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			val, curr := parsePrice(tc.input)
			// check with small delta for float comparison
			diff := val - tc.expectedVal
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 || curr != tc.expectedCurr {
				t.Errorf("parsePrice(%q) = (%f, %q); expected (%f, %q)", tc.input, val, curr, tc.expectedVal, tc.expectedCurr)
			}
		})
	}
}

func TestParseExternalId(t *testing.T) {
	tests := []struct {
		input    string
		expected *string
	}{
		{"https://www.ebay.com.au/itm/123", stringPtr("123")},
		{"https://www.ebay.com.au/itm/456", stringPtr("456")},
		{"https://example.com", nil},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			res := parseExternalId(tc.input)
			if (res == nil && tc.expected != nil) || (res != nil && tc.expected == nil) {
				t.Errorf("parseExternalId(%q) = %v; expected %v", tc.input, res, tc.expected)
			} else if res != nil && tc.expected != nil && *res != *tc.expected {
				t.Errorf("parseExternalId(%q) = %q; expected %q", tc.input, *res, *tc.expected)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
