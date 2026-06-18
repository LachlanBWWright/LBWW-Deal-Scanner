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

func TestBuildCcApiUrl(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		minPrice *float64
		maxPrice *float64
		expected string
	}{
		{
			name:     "raw query term",
			input:    "bananza",
			expected: "https://www.cashconverters.com.au/c3api/search/results?query=bananza",
		},
		{
			name:     "already API url",
			input:    "https://www.cashconverters.com.au/c3api/search/results?Sort=Default&page=1&SalePrice=20%7C99999999%7CC&query=bananza",
			expected: "https://www.cashconverters.com.au/c3api/search/results?Sort=Default&page=1&SalePrice=20%7C99999999%7CC&query=bananza",
		},
		{
			name:     "web search url",
			input:    "https://www.cashconverters.com.au/search-results?query=bananza&Sort=newest",
			expected: "https://www.cashconverters.com.au/c3api/search/results?query=bananza&Sort=newest",
		},
		{
			name:     "web search url with different casing and page",
			input:    "https://www.cashconverters.com.au/search-results?q=bananza&sort=price%2Cdesc&page=2",
			expected: "https://www.cashconverters.com.au/c3api/search/results?q=bananza&sort=price%2Cdesc&page=2",
		},
		{
			name:     "preserves every filter parameter",
			input:    "https://www.cashconverters.com.au/search-results?query=console&salePrice=150%7C900%7CC&category=gaming&store=123&customFilter=yes",
			expected: "https://www.cashconverters.com.au/c3api/search/results?query=console&salePrice=150%7C900%7CC&category=gaming&store=123&customFilter=yes",
		},
		{
			name:     "both prices override URL price",
			input:    "https://www.cashconverters.com.au/search-results?query=console&SalePrice%5B0%5D=20%7C500%7CC&category=gaming",
			minPrice: float64Pointer(50),
			maxPrice: float64Pointer(100),
			expected: "https://www.cashconverters.com.au/c3api/search/results?SalePrice%5B0%5D=50%7C100%7C&category=gaming&query=console",
		},
		{
			name:     "minimum price supplies default maximum",
			input:    "https://www.cashconverters.com.au/search-results?query=console",
			minPrice: float64Pointer(50),
			expected: "https://www.cashconverters.com.au/c3api/search/results?SalePrice%5B0%5D=50%7C999999%7C&query=console",
		},
		{
			name:     "maximum price supplies zero minimum",
			input:    "https://www.cashconverters.com.au/search-results?query=console",
			maxPrice: float64Pointer(100),
			expected: "https://www.cashconverters.com.au/c3api/search/results?SalePrice%5B0%5D=0%7C100%7C&query=console",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := buildCcApiUrl(tc.input, tc.minPrice, tc.maxPrice)
			if res != tc.expected {
				t.Errorf("buildCcApiUrl(%q) =\n%s\nexpected:\n%s", tc.input, res, tc.expected)
			}
		})
	}
}

func float64Pointer(value float64) *float64 {
	return &value
}
