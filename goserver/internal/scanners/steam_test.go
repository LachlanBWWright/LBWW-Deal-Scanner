package scanners

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSteamResponseParsing(t *testing.T) {
	rawJson := `{
		"results": [
			{
				"name": "AK-47 | Redline (Field-Tested)",
				"sell_price": 2500
			}
		]
	}`

	var resp SteamMarketResponse
	err := json.NewDecoder(strings.NewReader(rawJson)).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode mock Steam response: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(resp.Results))
	}

	item := resp.Results[0]
	if item.Name != "AK-47 | Redline (Field-Tested)" {
		t.Errorf("Expected name %q, got %q", "AK-47 | Redline (Field-Tested)", item.Name)
	}

	if item.SellPrice != 2500 {
		t.Errorf("Expected sell price 2500, got %f", item.SellPrice)
	}
}
