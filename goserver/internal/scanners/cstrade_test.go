package scanners

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCsTradeResponseParsing(t *testing.T) {
	rawJson := `{
		"inventory": [
			{
				"id": "12345",
				"app_id": 730,
				"market_hash_name": "AK-47 | Redline (Field-Tested)",
				"price": 25.50,
				"wear": 0.15,
				"icon": "icon_url",
				"status": "tradable"
			}
		]
	}`

	var resp CsTradeResponse
	err := json.NewDecoder(strings.NewReader(rawJson)).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode mock CsTrade response: %v", err)
	}

	if len(resp.Inventory) != 1 {
		t.Fatalf("Expected 1 inventory item, got %d", len(resp.Inventory))
	}

	item := resp.Inventory[0]
	if item.MarketHashName != "AK-47 | Redline (Field-Tested)" {
		t.Errorf("Expected MarketHashName %q, got %q", "AK-47 | Redline (Field-Tested)", item.MarketHashName)
	}
	if item.Price != 25.50 {
		t.Errorf("Expected price 25.50, got %f", item.Price)
	}
	if item.Wear != 0.15 {
		t.Errorf("Expected wear 0.15, got %f", item.Wear)
	}
}

func TestLootFarmResponseParsing(t *testing.T) {
	rawJson := `{
		"result": {
			"AK-47 | Redline (Field-Tested)": {
				"n": "AK-47 | Redline (Field-Tested)",
				"p": 2550,
				"u": {
					"1": [
						{
							"id": "999",
							"f": "15000",
							"st": null
						}
					]
				}
			}
		}
	}`

	var resp LootFarmResponse
	err := json.NewDecoder(strings.NewReader(rawJson)).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode mock LootFarm response: %v", err)
	}

	skin, ok := resp.Result["AK-47 | Redline (Field-Tested)"]
	if !ok {
		t.Fatalf("Expected skin in result map")
	}

	if skin.N != "AK-47 | Redline (Field-Tested)" {
		t.Errorf("Expected skin name %q, got %q", "AK-47 | Redline (Field-Tested)", skin.N)
	}
	if skin.P != 2550 {
		t.Errorf("Expected skin p 2550, got %f", skin.P)
	}

	items, ok := skin.U["1"]
	if !ok || len(items) != 1 {
		t.Fatalf("Expected 1 item for bot 1")
	}

	item := items[0]
	if item.ID != "999" {
		t.Errorf("Expected item ID 999, got %s", item.ID)
	}
	if item.F != "15000" {
		t.Errorf("Expected item float string '15000', got %s", item.F)
	}
}

func TestTradeItResponseParsing(t *testing.T) {
	rawJson := `{
		"items": [
			{
				"id": "555",
				"price": 2500,
				"name": "AK-47 | Redline (Field-Tested)",
				"floatValue": 0.15
			}
		]
	}`

	var resp TradeItResponse
	err := json.NewDecoder(strings.NewReader(rawJson)).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode mock TradeIt response: %v", err)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(resp.Items))
	}

	item := resp.Items[0]
	if item.Name != "AK-47 | Redline (Field-Tested)" {
		t.Errorf("Expected name %q, got %q", "AK-47 | Redline (Field-Tested)", item.Name)
	}
	if item.Price != 2500 {
		t.Errorf("Expected price 2500, got %f", item.Price)
	}
	if item.FloatValue == nil || *item.FloatValue != 0.15 {
		t.Errorf("Expected floatValue 0.15, got %v", item.FloatValue)
	}
}
