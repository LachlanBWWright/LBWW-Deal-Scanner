package scanners

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCsMarketJSONParsing(t *testing.T) {
	// Test the object-based ListingInfo format returned by Steam Community Market
	mockJSON := `{
		"listinginfo": {
			"123456789": {
				"listingid": "123456789",
				"converted_price_per_unit": 2500,
				"converted_fee_per_unit": 300,
				"asset": {
					"id": "987654321",
					"market_actions": [
						{
							"link": "steam://rungame/730/76561202255233023/+csgo_econ_action_preview%20M%listingid%A%assetid%D12345"
						}
					]
				}
			}
		}
	}`

	var raw struct {
		ListingInfo json.RawMessage `json:"listinginfo"`
	}
	err := json.Unmarshal([]byte(mockJSON), &raw)
	if err != nil {
		t.Fatalf("Failed to parse mock JSON: %v", err)
	}

	res := &SteamCsMarketResponse{
		ListingInfo: make(map[string]SteamListing),
	}

	if len(raw.ListingInfo) > 0 && raw.ListingInfo[0] == '{' {
		err = json.Unmarshal(raw.ListingInfo, &res.ListingInfo)
		if err != nil {
			t.Fatalf("Failed to unmarshal ListingInfo: %v", err)
		}
	} else {
		t.Fatalf("ListingInfo is not an object as expected")
	}

	if len(res.ListingInfo) != 1 {
		t.Fatalf("Expected 1 listing, got %d", len(res.ListingInfo))
	}

	listing, ok := res.ListingInfo["123456789"]
	if !ok {
		t.Fatalf("Listing '123456789' not found in parsed map")
	}

	if listing.ListingID != "123456789" {
		t.Errorf("Expected ListingID '123456789', got %q", listing.ListingID)
	}

	if listing.ConvertedPricePerUnit != 2500 {
		t.Errorf("Expected price 2500, got %f", listing.ConvertedPricePerUnit)
	}

	if len(listing.Asset.MarketActions) != 1 {
		t.Fatalf("Expected 1 market action, got %d", len(listing.Asset.MarketActions))
	}

	action := listing.Asset.MarketActions[0]
	if !strings.Contains(action.Link, "%listingid%") {
		t.Errorf("Expected link to contain placeholder %%listingid%%, got %q", action.Link)
	}
}

func TestCsMarketLiveIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live integration test in short mode")
	}

	// Use a real public CS2 market item endpoint to verify connection and schema
	// This item is usually present (e.g. Recoil Case)
	liveURL := "https://steamcommunity.com/market/listings/730/Recoil%20Case/render/?query=&start=0&count=3&country=US&language=english&currency=1"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", liveURL, nil)
	if err != nil {
		t.Fatalf("Failed to build request: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute live HTTP request: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Steam live response status: %s", resp.Status)

	if resp.StatusCode == http.StatusTooManyRequests {
		t.Skip("Skipping test due to Steam rate-limiting (429 Too Many Requests)")
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Steam returned non-OK status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body bytes: %v", err)
	}

	var raw struct {
		ListingInfo json.RawMessage `json:"listinginfo"`
	}
	err = json.Unmarshal(bodyBytes, &raw)
	if err != nil {
		if len(bodyBytes) > 0 && bodyBytes[0] == '<' {
			t.Skip("Skipping test: Steam returned HTML (likely captcha/auth block page) instead of JSON")
		}
		t.Fatalf("Failed to decode live response body: %v. Response body: %s", err, string(bodyBytes[:min(len(bodyBytes), 500)]))
	}

	res := &SteamCsMarketResponse{
		ListingInfo: make(map[string]SteamListing),
	}

	if len(raw.ListingInfo) > 0 && raw.ListingInfo[0] == '{' {
		err = json.Unmarshal(raw.ListingInfo, &res.ListingInfo)
		if err != nil {
			t.Fatalf("Failed to parse live ListingInfo JSON: %v", err)
		}
		t.Logf("Successfully fetched and parsed %d live listings from Steam Market!", len(res.ListingInfo))
	} else if len(raw.ListingInfo) > 0 && raw.ListingInfo[0] == '[' {
		// Empty array representation when there are no items
		t.Log("Steam returned empty listing array (no active listings).")
	} else {
		t.Fatalf("Unknown/unexpected ListingInfo response format starting with byte: %c", raw.ListingInfo[0])
	}
}
