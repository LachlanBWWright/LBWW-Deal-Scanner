package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/notifications"

	"github.com/PuerkitoBio/goquery"
)

type CashConvertersScanner struct {
	dbClient *qry.Query
}

func NewCashConvertersScanner(dbClient *qry.Query) *CashConvertersScanner {
	return &CashConvertersScanner{dbClient: dbClient}
}

func (s *CashConvertersScanner) Name() string {
	return "Cash Converters"
}

type CcSummary struct {
	CanonicalUrl string
	Title        string
	Price        float64
	Shipping     float64
	TotalPrice   float64
	ImageUrl     string
}

type CcDetail struct {
	CcSummary
	Description  string
	Availability string
}

type CcApiProductItem struct {
	Title            string `json:"Title"`
	Sp               string `json:"Sp"`
	ShippingCost     string `json:"ShippingCost"`
	AbsoluteImageUrl string `json:"AbsoluteImageUrl"`
	Url              string `json:"Url"`
}

type CcApiResponse struct {
	WasSuccessful bool `json:"WasSuccessful"`
	Value         struct {
		ProductList struct {
			ProductListItems []CcApiProductItem `json:"ProductListItems"`
		} `json:"ProductList"`
	} `json:"Value"`
}

func buildCcApiUrl(searchUrl string) string {
	if !strings.HasPrefix(searchUrl, "http") {
		apiUrl := url.URL{
			Scheme: "https",
			Host:   "www.cashconverters.com.au",
			Path:   "/c3api/search/results",
		}
		newQ := url.Values{}
		newQ.Set("query", searchUrl)
		newQ.Set("Sort", "Default")
		newQ.Set("SalePrice", "20|99999999|C")
		newQ.Set("page", "1")
		apiUrl.RawQuery = newQ.Encode()
		return apiUrl.String()
	}

	if strings.Contains(searchUrl, "/c3api/search/results") {
		return searchUrl
	}

	parsed, err := url.Parse(searchUrl)
	if err != nil {
		return searchUrl
	}

	q := parsed.Query()
	queryVal := q.Get("query")
	if queryVal == "" {
		queryVal = q.Get("q")
	}

	sortVal := q.Get("Sort")
	if sortVal == "" {
		sortVal = q.Get("sort")
	}
	if sortVal == "" {
		sortVal = "Default"
	}

	salePriceVal := q.Get("SalePrice")
	if salePriceVal == "" {
		salePriceVal = q.Get("saleprice")
	}
	if salePriceVal == "" {
		salePriceVal = "20|99999999|C"
	}

	pageVal := q.Get("page")
	if pageVal == "" {
		pageVal = "1"
	}

	apiUrl := url.URL{
		Scheme: "https",
		Host:   "www.cashconverters.com.au",
		Path:   "/c3api/search/results",
	}
	newQ := url.Values{}
	newQ.Set("query", queryVal)
	newQ.Set("Sort", sortVal)
	newQ.Set("SalePrice", salePriceVal)
	newQ.Set("page", pageVal)
	apiUrl.RawQuery = newQ.Encode()

	return apiUrl.String()
}

func (s *CashConvertersScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	globals, err := s.dbClient.GetGlobals(ctx)
	if err != nil {
		return nil, err
	}
	if globals == nil || !globals.CashConverters {
		return nil, nil
	}

	queries, err := s.dbClient.ListSavedQueries(ctx, "cashConverters")
	if err != nil {
		return nil, err
	}

	var notifs []notifications.AppNotification
	now := time.Now().UTC()

	for _, query := range queries {
		if query.Url == nil {
			continue
		}

		time.Sleep(3 * time.Second) // protect against rate limits

		summaries, err := s.discoverCcItems(ctx, *query.Url)
		if err != nil {
			log.Printf("Cash Converters API fetch failed for URL %s: %v", *query.Url, err)
			continue
		}

		for _, sum := range summaries {
			needDescription := (query.RequiredInDescription != nil && *query.RequiredInDescription != "") ||
				(query.ExcludeInDescription != nil && *query.ExcludeInDescription != "")

			var detail CcDetail
			if needDescription {
				detail, err = s.getCcDetail(ctx, sum)
				if err != nil {
					log.Printf("Failed to get product details for %s: %v", sum.CanonicalUrl, err)
					continue
				}
			} else {
				detail = CcDetail{
					CcSummary:    sum,
					Description:  "",
					Availability: "available",
				}
			}

			// Persist listing observation
			found := qry.DiscoveredListing{
				Source:       "cashConverters",
				CanonicalUrl: detail.CanonicalUrl,
				Title:        detail.Title,
				Price:        &detail.Price,
				Shipping:     &detail.Shipping,
				TotalPrice:   &detail.TotalPrice,
				ImageUrl:     &detail.ImageUrl,
				Description:  &detail.Description,
				Availability: &detail.Availability,
			}
			listingId := qry.StableListingId("cashConverters", detail.CanonicalUrl)
			_, err = s.dbClient.PersistListingObservation(ctx, found, now)
			if err != nil {
				continue
			}

			// Evaluate keyword and description filters
			matched := true
			var rejectReason *string

			searchable := strings.ToLower(detail.Title + " " + detail.Description)
			if query.RequiredPhrases != nil && *query.RequiredPhrases != "" {
				for _, phrase := range parsePhrases(*query.RequiredPhrases) {
					if !strings.Contains(searchable, phrase) {
						matched = false
						reason := "MissingRequiredPhrase"
						rejectReason = &reason
						break
					}
				}
			}

			if matched && query.ExcludePhrases != nil && *query.ExcludePhrases != "" {
				for _, phrase := range parsePhrases(*query.ExcludePhrases) {
					if strings.Contains(searchable, phrase) {
						matched = false
						reason := "ExcludedPhrase"
						rejectReason = &reason
						break
					}
				}
			}

			descSearchable := strings.ToLower(detail.Description)
			if matched && query.RequiredInDescription != nil && *query.RequiredInDescription != "" {
				for _, phrase := range parsePhrases(*query.RequiredInDescription) {
					if !strings.Contains(descSearchable, phrase) {
						matched = false
						reason := "MissingRequiredInDescription"
						rejectReason = &reason
						break
					}
				}
			}

			if matched && query.ExcludeInDescription != nil && *query.ExcludeInDescription != "" {
				for _, phrase := range parsePhrases(*query.ExcludeInDescription) {
					if strings.Contains(descSearchable, phrase) {
						matched = false
						reason := "ExcludedFromDescription"
						rejectReason = &reason
						break
					}
				}
			}

			if matched && query.MaxPrice != nil && detail.TotalPrice > *query.MaxPrice {
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

			prev, err := s.dbClient.GetQueryListingState(ctx, query.QueryId, listingId)
			if err != nil {
				continue
			}

			shouldNotify := false
			if matched {
				if prev == nil || prev.Status == "rejected" || prev.Status == "unseen" {
					shouldNotify = true
				}
			}

			state := &models.QueryListingState{
				QueryId:            query.QueryId,
				ListingId:          listingId,
				Source:             "cashConverters",
				Status:             status,
				LastEvaluatedAt:    now,
				LastRejectedReason: rejectReason,
			}

			if prev != nil {
				state.FirstMatchedAt = prev.FirstMatchedAt
				state.LowestObservedPrice = prev.LowestObservedPrice
				if prev.LowestObservedPrice == nil || detail.TotalPrice < *prev.LowestObservedPrice {
					state.LowestObservedPrice = &detail.TotalPrice
				}
			} else {
				state.LowestObservedPrice = &detail.TotalPrice
				if matched {
					state.FirstMatchedAt = &now
				}
			}

			if shouldNotify {
				status = "notified"
				state.Status = status
				state.LastNotifiedAt = &now
				state.LastNotifiedTotalPrice = &detail.TotalPrice

				title := fmt.Sprintf("a %s for $%.2f is available at %s", detail.Title, detail.TotalPrice, detail.CanonicalUrl)
				notifs = append(notifs, notifications.AppNotification{
					Kind:     "deal",
					Source:   "cashConverters",
					Title:    title,
					Url:      detail.CanonicalUrl,
					Price:    &detail.TotalPrice,
					ImageUrl: &detail.ImageUrl,
					Query: &notifications.NotificationQuery{
						Type: "cashConverters",
						Id:   query.Id,
					},
				})
			}

			s.dbClient.UpsertQueryListingState(ctx, state)
		}
	}

	return notifs, nil
}

func (s *CashConvertersScanner) discoverCcItems(ctx context.Context, searchUrl string) ([]CcSummary, error) {
	apiUrl := buildCcApiUrl(searchUrl)

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
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
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var apiResp CcApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !apiResp.WasSuccessful {
		return nil, fmt.Errorf("API response indicates failure")
	}

	var results []CcSummary
	for _, item := range apiResp.Value.ProductList.ProductListItems {
		price := parseCcPrice(item.Sp)
		shipping := parseCcShipping(item.ShippingCost)
		total := price + shipping

		href := item.Url
		if !strings.HasPrefix(href, "http") {
			href = "https://www.cashconverters.com.au" + href
		}

		results = append(results, CcSummary{
			CanonicalUrl: href,
			Title:        strings.Join(strings.Fields(item.Title), " "),
			Price:        price,
			Shipping:     shipping,
			TotalPrice:   total,
			ImageUrl:     item.AbsoluteImageUrl,
		})
	}

	return results, nil
}

func (s *CashConvertersScanner) getCcDetail(ctx context.Context, sum CcSummary) (CcDetail, error) {
	// Look up listing in DB
	id := qry.StableListingId("cashConverters", sum.CanonicalUrl)
	var listing models.Listing
	listingPtr, err := s.dbClient.Listing.WithContext(ctx).Where(s.dbClient.Listing.ID.Eq(id)).First()
	if err == nil {
		listing = *listingPtr
	}

	// If detail exists and is fresh (less than 24 hours), reuse it
	if err == nil && listing.LastDetailAt != nil && time.Since(*listing.LastDetailAt) < 24*time.Hour {
		desc := ""
		if listing.Description != nil {
			desc = *listing.Description
		}
		avail := "available"
		if listing.Availability != nil {
			avail = *listing.Availability
		}
		img := sum.ImageUrl
		if listing.ImageUrl != nil {
			img = *listing.ImageUrl
		}
		return CcDetail{
			CcSummary: CcSummary{
				CanonicalUrl: sum.CanonicalUrl,
				Title:        listing.Title,
				Price:        sum.Price,
				Shipping:     sum.Shipping,
				TotalPrice:   sum.TotalPrice,
				ImageUrl:     img,
			},
			Description:  desc,
			Availability: avail,
		}, nil
	}

	// Fetch fresh page detail
	time.Sleep(2 * time.Second)

	req, err := http.NewRequestWithContext(ctx, "GET", sum.CanonicalUrl, nil)
	if err != nil {
		return CcDetail{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return CcDetail{}, err
	}
	if resp == nil {
		return CcDetail{}, fmt.Errorf("http response is nil")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CcDetail{}, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return CcDetail{}, err
	}

	title := CleanSelectionText(doc.Find("h1, .product-detail__title, .product-title").First())
	if title == "" {
		title = strings.Join(strings.Fields(sum.Title), " ")
	}

	var description string
	// Try standard meta description tag
	doc.Find("meta[name='description']").Each(func(i int, sel *goquery.Selection) {
		if content, exists := sel.Attr("content"); exists {
			description = strings.TrimSpace(content)
		}
	})
	// Fall back to og:description
	if description == "" {
		doc.Find("meta[property='og:description']").Each(func(i int, sel *goquery.Selection) {
			if content, exists := sel.Attr("content"); exists {
				description = strings.TrimSpace(content)
			}
		})
	}
	// Fall back to selector description
	if description == "" {
		description = strings.TrimSpace(doc.Find(".product-detail__description, .product-description, [class*='description']").First().Text())
	}

	regSpace := regexp.MustCompile(`\s+`)
	description = regSpace.ReplaceAllString(description, " ")

	bodyText := strings.ToLower(doc.Find("body").Text())
	avail := "available"
	if strings.Contains(bodyText, "sold") {
		avail = "unavailable"
	}

	// Update last detail time in DB Listing
	now := time.Now().UTC()
	s.dbClient.Listing.WithContext(ctx).Where(s.dbClient.Listing.ID.Eq(id)).Update(s.dbClient.Listing.LastDetailAt, &now)

	return CcDetail{
		CcSummary: CcSummary{
			CanonicalUrl: sum.CanonicalUrl,
			Title:        title,
			Price:        sum.Price,
			Shipping:     sum.Shipping,
			TotalPrice:   sum.TotalPrice,
			ImageUrl:     sum.ImageUrl,
		},
		Description:  description,
		Availability: avail,
	}, nil
}

func parseCcPrice(input string) float64 {
	reg := regexp.MustCompile(`[^0-9.]`)
	clean := reg.ReplaceAllString(input, "")
	val, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0
	}
	return val
}

func parseCcShipping(input string) float64 {
	if strings.Contains(strings.ToLower(input), "free") {
		return 0
	}
	return parseCcPrice(input)
}

func parsePhrases(input string) []string {
	var results []string
	for _, p := range strings.Split(input, ",") {
		clean := strings.TrimSpace(strings.ToLower(p))
		if clean != "" {
			results = append(results, clean)
		}
	}
	return results
}
