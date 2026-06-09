package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"dealscanner/internal/models"
	qry "dealscanner/internal/db/query"
	"dealscanner/internal/notifications"
)

type CsMarketScanner struct {
	dbClient   *qry.Query
	mu         sync.Mutex
	index      int
	itemsFound map[string]int
}

func NewCsMarketScanner(dbClient *qry.Query) *CsMarketScanner {
	return &CsMarketScanner{
		dbClient:   dbClient,
		itemsFound: make(map[string]int),
	}
}

func (s *CsMarketScanner) Name() string {
	return "CS Market"
}

type SteamListingAssetAction struct {
	Link string `json:"link"`
}

type SteamListingAsset struct {
	ID            string                     `json:"id"`
	MarketActions []SteamListingAssetAction `json:"market_actions"`
}

type SteamListing struct {
	ListingID             string            `json:"listingid"`
	ConvertedPricePerUnit float64           `json:"converted_price_per_unit"`
	ConvertedFeePerUnit   float64           `json:"converted_fee_per_unit"`
	Asset                 SteamListingAsset `json:"asset"`
}

type SteamCsMarketResponse struct {
	ListingInfo map[string]SteamListing `json:"listinginfo"`
}

type SteamItemInfo struct {
	FloatValue   float64 `json:"floatvalue"`
	FullItemName string  `json:"full_item_name"`
}

type SteamItemInfoResponse struct {
	ItemInfo SteamItemInfo `json:"iteminfo"`
}

func (s *CsMarketScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	globals, err := s.dbClient.GetGlobals(ctx)
	if err != nil {
		return nil, err
	}
	if globals == nil || !globals.CsItems {
		return nil, nil
	}

	results, err := s.dbClient.CsMarket.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}

	s.mu.Lock()
	if s.index >= len(results) {
		s.index = 0
	}
	item := *results[s.index]
	s.index++
	s.mu.Unlock()

	time.Sleep(3 * time.Second) // protect rate limits

	listingData, err := s.fetchSteamCsMarketListing(ctx, item.Url)
	if err != nil {
		log.Printf("Failed to fetch CS market listing: %v", err)
		return nil, err
	}

	var notifs []notifications.AppNotification
	s.processCsMarketListings(ctx, item, listingData, &notifs)

	return notifs, nil
}

func (s *CsMarketScanner) fetchSteamCsMarketListing(ctx context.Context, itemUrl string) (*SteamCsMarketResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", itemUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("http response is nil")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam returned status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Deal with the array vs object mismatch for listinginfo
	var raw struct {
		ListingInfo json.RawMessage `json:"listinginfo"`
	}
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, err
	}

	res := &SteamCsMarketResponse{
		ListingInfo: make(map[string]SteamListing),
	}

	if len(raw.ListingInfo) > 0 && raw.ListingInfo[0] == '{' {
		if err := json.Unmarshal(raw.ListingInfo, &res.ListingInfo); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (s *CsMarketScanner) processCsMarketListings(ctx context.Context, item models.CsMarket, listingData *SteamCsMarketResponse, notifs *[]notifications.AppNotification) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := 0
	for _, listing := range listingData.ListingInfo {
		if len(listing.Asset.MarketActions) == 0 {
			continue
		}
		action := listing.Asset.MarketActions[0]
		query := fmt.Sprintf("https://api.csgofloat.com/?url=%s", action.Link)
		query = strings.ReplaceAll(query, "%listingid%", listing.ListingID)
		query = strings.ReplaceAll(query, "%assetid%", listing.Asset.ID)

		price := (listing.ConvertedPricePerUnit + listing.ConvertedFeePerUnit) / 100.0

		_, found := s.itemsFound[query]
		if !found && i < 10 {
			time.Sleep(1 * time.Second) // be nice to csgofloat
			itemInfo, err := s.fetchSteamItemInfo(ctx, query)
			if err != nil {
				log.Printf("Failed to fetch item info for query %s: %v", query, err)
				i++
				continue
			}

			if itemInfo.ItemInfo.FloatValue < item.MaxFloat && price <= item.MaxPrice {
				title := fmt.Sprintf("a %s with float %.5f is available for $%.2f USD at: %s", itemInfo.ItemInfo.FullItemName, itemInfo.ItemInfo.FloatValue, price, item.DisplayUrl)
				*notifs = append(*notifs, notifications.AppNotification{
					Kind:     "deal",
					Source:   "steamMarket",
					Title:    title,
					Url:      item.DisplayUrl,
					Price:    &price,
					ImageUrl: nil,
					Query: &notifications.NotificationQuery{
						Type: "csMarket",
						Id:   item.Url,
					},
				})
			}
		}

		if i < 10 || found {
			s.itemsFound[query] = 20
		}
		i++
	}

	// Decrement counts and delete expired ones
	for k, v := range s.itemsFound {
		newVal := v - 1
		if newVal <= 0 {
			delete(s.itemsFound, k)
		} else {
			s.itemsFound[k] = newVal
		}
	}
}

func (s *CsMarketScanner) fetchSteamItemInfo(ctx context.Context, itemUrl string) (*SteamItemInfoResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", itemUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("http response is nil")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("csgofloat returned status: %d", resp.StatusCode)
	}

	var res SteamItemInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}
