package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/notifications"
)

type SteamMarketScanner struct {
	dbClient *qry.Query
}

func NewSteamMarketScanner(dbClient *qry.Query) *SteamMarketScanner {
	return &SteamMarketScanner{dbClient: dbClient}
}

func (s *SteamMarketScanner) Name() string {
	return "Steam Market"
}

type SteamMarketItem struct {
	Name      string  `json:"name"`
	SellPrice float64 `json:"sell_price"`
}

type SteamMarketResponse struct {
	Results []SteamMarketItem `json:"results"`
}

func (s *SteamMarketScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	// Load globals to check if enabled
	globals, err := s.dbClient.GetGlobals(ctx)
	if err != nil {
		return nil, err
	}
	if globals == nil || !globals.SteamQuery {
		return nil, nil
	}

	queries, err := s.dbClient.ListSavedQueries(ctx, "steamMarket")
	if err != nil {
		return nil, err
	}

	var notifs []notifications.AppNotification
	client := &http.Client{Timeout: 10 * time.Second}

	for _, item := range queries {
		if item.Name == nil {
			continue
		}

		time.Sleep(3 * time.Second) // Steam rate limits protection

		req, err := http.NewRequestWithContext(ctx, "GET", *item.Name, nil)
		if err != nil {
			log.Printf("Failed to create Steam request for %s: %v", *item.Name, err)
			continue
		}
		// Mimic browser user-agent
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Steam SCM request failed for %s: %v", *item.Name, err)
			continue
		}
		if resp == nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Steam SCM returned status: %d for %s", resp.StatusCode, *item.Name)
			resp.Body.Close()
			continue
		}

		var steamResponse SteamMarketResponse
		err = json.NewDecoder(resp.Body).Decode(&steamResponse)
		resp.Body.Close()
		if err != nil {
			log.Printf("Failed to decode Steam response for %s: %v", *item.Name, err)
			continue
		}

		var prices []float64
		for _, result := range steamResponse.Results {
			price := result.SellPrice / 100.0
			prices = append(prices, price)

			// Persist listing
			canonicalUrl := fmt.Sprintf("https://steamcommunity.com/market/listings/730/%s", result.Name)
			found := qry.DiscoveredListing{
				Source:       "steamMarket",
				CanonicalUrl: canonicalUrl,
				Title:        result.Name,
				Price:        &price,
				TotalPrice:   &price,
			}

			listingId := qry.StableListingId("steamMarket", canonicalUrl)
			_, err = s.dbClient.PersistListingObservation(ctx, found, time.Now().UTC())
			if err != nil {
				log.Printf("Failed to persist SCM observation: %v", err)
				continue
			}

			now := time.Now().UTC()
			match := qry.MatchPriceRange(price, nil, item.MaxPrice)
			decision, err := s.dbClient.EvaluateListingForQuery(ctx, item.QueryId, listingId, "steamMarket", &price, match, now, 0)
			if err != nil {
				return nil, err
			}

			if decision.Type == qry.DecisionTypeNotify {
				title := fmt.Sprintf("a %s is available for $%.2f USD at: %s", result.Name, price, canonicalUrl)
				notifs = append(notifs, notifications.AppNotification{
					Kind:     "deal",
					Source:   "steamMarket",
					Title:    title,
					Url:      canonicalUrl,
					Price:    &price,
					ImageUrl: nil,
					Query: &notifications.NotificationQuery{
						Type: "steamMarket",
						Id:   item.Id,
					},
				})
			}
		}

		if len(prices) > 0 {
			// Update last price on SteamMarket query
			lowestPrice := prices[0]
			for _, p := range prices {
				if p < lowestPrice {
					lowestPrice = p
				}
			}

			_, err := s.dbClient.SteamMarket.WithContext(ctx).Where(s.dbClient.SteamMarket.Name.Eq(*item.Name)).Update(s.dbClient.SteamMarket.LastPrice, lowestPrice)
			if err != nil {
				return nil, err
			}
		}
	}

	return notifs, nil
}
