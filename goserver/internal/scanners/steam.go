package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"dealscanner/internal/db"
	"dealscanner/internal/notifications"
)

type SteamMarketScanner struct {
	dbClient *db.DB
}

func NewSteamMarketScanner(dbClient *db.DB) *SteamMarketScanner {
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
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Steam SCM returned status: %d for %s", resp.StatusCode, *item.Name)
			continue
		}

		var steamResponse SteamMarketResponse
		if err := json.NewDecoder(resp.Body).Decode(&steamResponse); err != nil {
			log.Printf("Failed to decode Steam response for %s: %v", *item.Name, err)
			continue
		}

		var prices []float64
		for _, result := range steamResponse.Results {
			price := result.SellPrice / 100.0
			prices = append(prices, price)

			// Persist listing
			canonicalUrl := fmt.Sprintf("https://steamcommunity.com/market/listings/730/%s", result.Name)
			found := db.DiscoveredListing{
				Source:       "steamMarket",
				CanonicalUrl: canonicalUrl,
				Title:        result.Name,
				Price:        &price,
				TotalPrice:   &price,
			}

			listingId := db.StableListingId("steamMarket", canonicalUrl)
			_, err = s.dbClient.PersistListingObservation(ctx, found, time.Now().UTC())
			if err != nil {
				log.Printf("Failed to persist SCM observation: %v", err)
				continue
			}

			// Evaluate
			matched := true
			var rejectReason *string

			if item.MaxPrice != nil && price > *item.MaxPrice {
				matched = false
				reason := "AboveMaxPrice"
				rejectReason = &reason
			}

			var status string
			if matched {
				status = "matched"
			} else {
				status = "rejected"
			}

			prev, err := s.dbClient.GetQueryListingState(ctx, item.Id, listingId)
			if err != nil {
				continue
			}

			shouldNotify := false
			if matched {
				if prev == nil || prev.Status == "rejected" || prev.Status == "unseen" {
					shouldNotify = true
				}
			}

			now := time.Now().UTC()
			state := &db.QueryListingState{
				QueryId:            item.Id,
				ListingId:          listingId,
				Source:             "steamMarket",
				Status:             status,
				LastEvaluatedAt:    now,
				LastRejectedReason: rejectReason,
			}

			if prev != nil {
				state.FirstMatchedAt = prev.FirstMatchedAt
				state.LowestObservedPrice = prev.LowestObservedPrice
				if prev.LowestObservedPrice == nil || price < *prev.LowestObservedPrice {
					state.LowestObservedPrice = &price
				}
			} else {
				state.LowestObservedPrice = &price
				if matched {
					state.FirstMatchedAt = &now
				}
			}

			if shouldNotify {
				status = "notified"
				state.Status = status
				state.LastNotifiedAt = &now
				state.LastNotifiedTotalPrice = &price

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

			s.dbClient.UpsertQueryListingState(ctx, state)
		}

		if len(prices) > 0 {
			// Update last price on SteamMarket query
			lowestPrice := prices[0]
			for _, p := range prices {
				if p < lowestPrice {
					lowestPrice = p
				}
			}

			s.dbClient.WithContext(ctx).Model(&db.SteamMarket{}).Where("name = ?", *item.Name).Update("lastPrice", lowestPrice)
		}
	}

	return notifs, nil
}
