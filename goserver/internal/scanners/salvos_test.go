package scanners

import (
	"testing"
)

func TestParseSalvosPrice(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"$49.99", 49.99},
		{"Price: $5.00", 5.00},
		{"120.0", 120.0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			val, err := parseSalvosPrice(tc.input)
			if err != nil {
				t.Fatalf("parseSalvosPrice(%q) failed: %v", tc.input, err)
			}
			diff := val - tc.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("parseSalvosPrice(%q) = %f; expected %f", tc.input, val, tc.expected)
			}
		})
	}
}
