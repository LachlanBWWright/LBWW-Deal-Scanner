package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/notifications"
)

type CsTradeScanner struct {
	dbClient *qry.Query
}

func NewCsTradeScanner(dbClient *qry.Query) *CsTradeScanner {
	return &CsTradeScanner{dbClient: dbClient}
}

func (s *CsTradeScanner) Name() string {
	return "CS Trade"
}

type CsTradeItem struct {
	ID             string  `json:"id"`
	AppID          int     `json:"app_id"`
	MarketHashName string  `json:"market_hash_name"`
	Price          float64 `json:"price"`
	Wear           float64 `json:"wear"`
	Icon           string  `json:"icon"`
	Status         string  `json:"status"`
}

type CsTradeResponse struct {
	Inventory []CsTradeItem `json:"inventory"`
}

func (s *CsTradeScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	globals, err := s.dbClient.GetGlobals(ctx)
	if err != nil {
		return nil, err
	}
	if globals == nil || !globals.CsItems {
		return nil, nil
	}

	queries, err := s.dbClient.ListSavedQueries(ctx, "csTradeBot")
	if err != nil {
		return nil, err
	}
	if len(queries) == 0 {
		return nil, nil
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get("https://cdn.cs.trade:8443/api/getInventory?order_by=price_desc&bot=all&_=1651756783463")
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("http response is nil")
	}
	defer resp.Body.Close()

	var data CsTradeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var notifs []notifications.AppNotification
	now := time.Now().UTC()

	for _, query := range queries {
		if query.Name == nil {
			continue
		}

		for _, item := range data.Inventory {
			if item.AppID != 730 || item.MarketHashName != *query.Name {
				continue
			}

			if item.Price > *query.MaxPrice || item.Wear < *query.MinFloat || item.Wear > *query.MaxFloat {
				continue
			}

			// Persist
			canonicalUrl := fmt.Sprintf("https://cs.trade/item/%s", item.ID)
			found := qry.DiscoveredListing{
				Source:       "csTrade",
				ExternalId:   &item.ID,
				CanonicalUrl: canonicalUrl,
				Title:        item.MarketHashName,
				Price:        &item.Price,
				TotalPrice:   &item.Price,
				ImageUrl:     &item.Icon,
				Availability: &item.Status,
			}
			listingId := qry.StableListingId("csTrade", canonicalUrl)
			_, err = s.dbClient.PersistListingObservation(ctx, found, now)
			if err != nil {
				continue
			}

			prev, err := s.dbClient.GetQueryListingState(ctx, query.QueryId, listingId)
			if err != nil {
				continue
			}

			if prev == nil || prev.Status == "rejected" || prev.Status == "unseen" {
				// Notify
				title := fmt.Sprintf("a %s with a float of %.5f is available for $%.2f USD at: https://cs.trade/", item.MarketHashName, item.Wear, item.Price)
				notifs = append(notifs, notifications.AppNotification{
					Kind:     "deal",
					Source:   "csTrade",
					Title:    title,
					Url:      "https://cs.trade/",
					Price:    &item.Price,
					ImageUrl: &item.Icon,
					Query: &notifications.NotificationQuery{
						Type: "csTradeBot",
						Id:   query.Id,
					},
				})

				state := &models.QueryListingState{
					QueryId:                query.QueryId,
					ListingId:              listingId,
					Source:                 "csTrade",
					Status:                 "notified",
					LastEvaluatedAt:        now,
					FirstMatchedAt:         &now,
					LastMatchedAt:          &now,
					LastNotifiedAt:         &now,
					LastNotifiedTotalPrice: &item.Price,
					LowestObservedPrice:    &item.Price,
				}
				s.dbClient.UpsertQueryListingState(ctx, state)
			}
		}
	}

	return notifs, nil
}

type LootFarmScanner struct {
	dbClient *qry.Query
}

func NewLootFarmScanner(dbClient *qry.Query) *LootFarmScanner {
	return &LootFarmScanner{dbClient: dbClient}
}

func (s *LootFarmScanner) Name() string {
	return "Loot Farm"
}

type LootFarmItem struct {
	ID string      `json:"id"`
	F  string      `json:"f"`
	ST interface{} `json:"st"`
}

type LootFarmSkin struct {
	N string                    `json:"n"`
	P float64                   `json:"p"`
	U map[string][]LootFarmItem `json:"u"`
}

type LootFarmResponse struct {
	Result map[string]LootFarmSkin `json:"result"`
}

func (s *LootFarmScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	globals, err := s.dbClient.GetGlobals(ctx)
	if err != nil {
		return nil, err
	}
	if globals == nil || !globals.CsItems {
		return nil, nil
	}

	queries, err := s.dbClient.ListSavedQueries(ctx, "csTradeBot")
	if err != nil {
		return nil, err
	}
	if len(queries) == 0 {
		return nil, nil
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get("https://loot.farm/botsInventory_730.json")
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("http response is nil")
	}
	defer resp.Body.Close()

	var data LootFarmResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var notifs []notifications.AppNotification
	now := time.Now().UTC()

	for _, query := range queries {
		if query.Name == nil {
			continue
		}

		for _, skin := range data.Result {
			if !strings.Contains(*query.Name, skin.N) {
				continue
			}

			price := skin.P / 100.0
			if price > *query.MaxPrice {
				continue
			}

			// Iterate bots inventory
			for botNo, items := range skin.U {
				_ = botNo
				for _, item := range items {
					if item.F == "" {
						continue
					}
					var fVal float64
					fmt.Sscanf(item.F, "%f", &fVal)
					itemFloat := fVal / 100000.0

					meetsFloat := itemFloat >= *query.MinFloat && itemFloat <= *query.MaxFloat
					meetsSt := !strings.Contains(*query.Name, "StatTrak") || item.ST != nil

					if !meetsFloat || !meetsSt {
						continue
					}

					canonicalUrl := fmt.Sprintf("https://loot.farm/item/%s", item.ID)
					found := qry.DiscoveredListing{
						Source:       "lootFarm",
						ExternalId:   &item.ID,
						CanonicalUrl: canonicalUrl,
						Title:        skin.N,
						Price:        &price,
						TotalPrice:   &price,
					}
					listingId := qry.StableListingId("lootFarm", canonicalUrl)
					_, err = s.dbClient.PersistListingObservation(ctx, found, now)
					if err != nil {
						continue
					}

					prev, err := s.dbClient.GetQueryListingState(ctx, query.QueryId, listingId)
					if err != nil {
						continue
					}

					if prev == nil || prev.Status == "rejected" || prev.Status == "unseen" {
						title := fmt.Sprintf("a %s with a float of %.5f is available for $%.2f USD at: https://loot.farm/", skin.N, itemFloat, price)
						notifs = append(notifs, notifications.AppNotification{
							Kind:     "deal",
							Source:   "lootFarm",
							Title:    title,
							Url:      "https://loot.farm/",
							Price:    &price,
							ImageUrl: nil,
							Query: &notifications.NotificationQuery{
								Type: "csTradeBot",
								Id:   query.Id,
							},
						})

						state := &models.QueryListingState{
							QueryId:                query.QueryId,
							ListingId:              listingId,
							Source:                 "lootFarm",
							Status:                 "notified",
							LastEvaluatedAt:        now,
							FirstMatchedAt:         &now,
							LastMatchedAt:          &now,
							LastNotifiedAt:         &now,
							LastNotifiedTotalPrice: &price,
							LowestObservedPrice:    &price,
						}
						s.dbClient.UpsertQueryListingState(ctx, state)
					}
				}
			}
		}
	}

	return notifs, nil
}

type TradeItScanner struct {
	dbClient *qry.Query
}

func NewTradeItScanner(dbClient *qry.Query) *TradeItScanner {
	return &TradeItScanner{dbClient: dbClient}
}

func (s *TradeItScanner) Name() string {
	return "Trade It"
}

type TradeItItem struct {
	ID          string    `json:"id"`
	Price       float64   `json:"price"`
	Name        string    `json:"name"`
	FloatValue  *float64  `json:"floatValue"`
	FloatValues []float64 `json:"floatValues"`
}

type TradeItResponse struct {
	Items []TradeItItem `json:"items"`
}

func (s *TradeItScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	globals, err := s.dbClient.GetGlobals(ctx)
	if err != nil {
		return nil, err
	}
	if globals == nil || !globals.CsItems {
		return nil, nil
	}

	queries, err := s.dbClient.ListSavedQueries(ctx, "csTradeBot")
	if err != nil {
		return nil, err
	}
	if len(queries) == 0 {
		return nil, nil
	}

	var allItems []TradeItItem
	client := &http.Client{Timeout: 15 * time.Second}

	for i := 0; i < 20; i++ {
		offset := i * 1000
		url := fmt.Sprintf("https://tradeit.gg/api/v2/inventory/data?gameId=730&offset=%d&limit=1000&sortType=(CSGO)+Best+Float&searchValue=&minPrice=0&maxPrice=100000&minFloat=0&maxFloat=1&hideTradeLock=false&fresh=true", offset)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			break
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			break
		}
		if resp == nil {
			break
		}

		var batchData TradeItResponse
		json.NewDecoder(resp.Body).Decode(&batchData)
		resp.Body.Close()

		allItems = append(allItems, batchData.Items...)
		if len(batchData.Items) < 750 {
			break
		}
	}

	var notifs []notifications.AppNotification
	now := time.Now().UTC()

	for _, query := range queries {
		if query.Name == nil {
			continue
		}

		for _, item := range allItems {
			if item.Name != *query.Name {
				continue
			}

			price := item.Price / 100.0
			if price > *query.MaxPrice {
				continue
			}

			bestFloat := 1.0
			if item.FloatValue != nil {
				bestFloat = *item.FloatValue
			} else if len(item.FloatValues) > 0 {
				bestFloat = item.FloatValues[0]
				for _, f := range item.FloatValues {
					if f < bestFloat {
						bestFloat = f
					}
				}
			}

			if bestFloat < *query.MinFloat || bestFloat > *query.MaxFloat {
				continue
			}

			canonicalUrl := fmt.Sprintf("https://tradeit.gg/item/%s", item.ID)
			found := qry.DiscoveredListing{
				Source:       "tradeIt",
				ExternalId:   &item.ID,
				CanonicalUrl: canonicalUrl,
				Title:        item.Name,
				Price:        &price,
				TotalPrice:   &price,
			}
			listingId := qry.StableListingId("tradeIt", canonicalUrl)
			_, err = s.dbClient.PersistListingObservation(ctx, found, now)
			if err != nil {
				continue
			}

			prev, err := s.dbClient.GetQueryListingState(ctx, query.QueryId, listingId)
			if err != nil {
				continue
			}

			if prev == nil || prev.Status == "rejected" || prev.Status == "unseen" {
				title := fmt.Sprintf("a %s with a float of %.5f is available for $%.2f USD at: https://tradeit.gg/csgo/trade", item.Name, bestFloat, price)
				notifs = append(notifs, notifications.AppNotification{
					Kind:     "deal",
					Source:   "tradeIt",
					Title:    title,
					Url:      "https://tradeit.gg/csgo/trade",
					Price:    &price,
					ImageUrl: nil,
					Query: &notifications.NotificationQuery{
						Type: "csTradeBot",
						Id:   query.Id,
					},
				})

				state := &models.QueryListingState{
					QueryId:                query.QueryId,
					ListingId:              listingId,
					Source:                 "tradeIt",
					Status:                 "notified",
					LastEvaluatedAt:        now,
					FirstMatchedAt:         &now,
					LastMatchedAt:          &now,
					LastNotifiedAt:         &now,
					LastNotifiedTotalPrice: &price,
					LowestObservedPrice:    &price,
				}
				s.dbClient.UpsertQueryListingState(ctx, state)
			}
		}
	}

	return notifs, nil
}
