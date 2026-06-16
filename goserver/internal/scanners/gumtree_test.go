package scanners

import (
	"testing"
)

func TestParseGumtreePrice(t *testing.T) {
	tests := []struct {
		input       string
		expectedVal float64
		expectErr   bool
	}{
		{"$10.00", 10.00, false},
		{"Free", 0.0, false},
		{"FREE", 0.0, false},
		{"$1,234.56", 1234.56, false},
		{"invalid", 0.0, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			val, err := parseGumtreePrice(tc.input)
			if (err != nil) != tc.expectErr {
				t.Errorf("parseGumtreePrice(%q) error state = %v; expected error = %v", tc.input, err != nil, tc.expectErr)
			}
			diff := val - tc.expectedVal
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("parseGumtreePrice(%q) value = %f; expected %f", tc.input, val, tc.expectedVal)
			}
		})
	}
}
