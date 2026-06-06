package scanners

import (
	"reflect"
	"testing"
)

func TestParseCcPrice(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"$123.45", 123.45},
		{"Price: $10.00", 10.00},
		{"15.5", 15.5},
		{"invalid", 0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			val := parseCcPrice(tc.input)
			diff := val - tc.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("parseCcPrice(%q) = %f; expected %f", tc.input, val, tc.expected)
			}
		})
	}
}

func TestParseCcShipping(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"FREE", 0},
		{"Free postage", 0},
		{"$5.00", 5.00},
		{"invalid", 0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			val := parseCcShipping(tc.input)
			diff := val - tc.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("parseCcShipping(%q) = %f; expected %f", tc.input, val, tc.expected)
			}
		})
	}
}

func TestParsePhrases(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{" xbox, Series X ,, Controller ", []string{"xbox", "series x", "controller"}},
		{"a, b, c", []string{"a", "b", "c"}},
		{"", nil},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			res := parsePhrases(tc.input)
			if len(res) == 0 && len(tc.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("parsePhrases(%q) = %v; expected %v", tc.input, res, tc.expected)
			}
		})
	}
}
