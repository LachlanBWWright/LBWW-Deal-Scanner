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
	client   *http.Client
}

func NewCashConvertersScanner(dbClient *qry.Query) *CashConvertersScanner {
	return &CashConvertersScanner{
		dbClient: dbClient,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
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

func buildCcApiUrl(searchUrl string, minPrice, maxPrice *float64) string {
	if !strings.HasPrefix(searchUrl, "http") {
		apiUrl := url.URL{
			Scheme: "https",
			Host:   "www.cashconverters.com.au",
			Path:   "/c3api/search/results",
		}
		q := url.Values{}
		q.Set("query", searchUrl)
		setCcPriceRange(q, minPrice, maxPrice)
		apiUrl.RawQuery = q.Encode()
		return apiUrl.String()
	}

	parsed, err := url.Parse(searchUrl)
	if err != nil {
		return searchUrl
	}

	parsed.Scheme = "https"
	parsed.Host = "www.cashconverters.com.au"
	parsed.Path = "/c3api/search/results"
	parsed.Fragment = ""
	if minPrice != nil || maxPrice != nil {
		q := parsed.Query()
		setCcPriceRange(q, minPrice, maxPrice)
		parsed.RawQuery = q.Encode()
	}
	return parsed.String()
}

func setCcPriceRange(q url.Values, minPrice, maxPrice *float64) {
	if minPrice == nil && maxPrice == nil {
		return
	}

	for key := range q {
		if strings.HasPrefix(strings.ToLower(key), "saleprice") {
			q.Del(key)
		}
	}

	min := 0.0
	if minPrice != nil {
		min = *minPrice
	}
	max := 999999.0
	if maxPrice != nil {
		max = *maxPrice
	}
	q.Set("SalePrice[0]", fmt.Sprintf("%g|%g|", min, max))
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
	descriptionFetchAvailable := true

	for _, query := range queries {
		if query.Url == nil {
			continue
		}

		summaries, err := s.discoverCcItems(ctx, *query.Url, query.MinPrice, query.MaxPrice)
		if err != nil {
			log.Printf("Cash Converters API fetch failed for URL %s: %v", *query.Url, err)
			continue
		}

		listingIDs := make([]string, 0, len(summaries))
		for _, summary := range summaries {
			listingIDs = append(listingIDs, qry.StableListingId("cashConverters", summary.CanonicalUrl))
		}
		existingListings, err := s.dbClient.LoadListings(ctx, listingIDs)
		if err != nil {
			log.Printf("Failed to load Cash Converters listings for URL %s: %v", *query.Url, err)
			continue
		}
		existingStates, err := s.dbClient.LoadQueryListingStates(ctx, query.QueryId, listingIDs)
		if err != nil {
			log.Printf("Failed to load Cash Converters states for URL %s: %v", *query.Url, err)
			continue
		}

		discovered := make([]qry.DiscoveredListing, 0, len(summaries))
		states := make([]*models.QueryListingState, 0, len(summaries))
		for _, sum := range summaries {
			requiredPhrases := combinePhraseFilters(query.RequiredPhrases, query.RequiredInDescription)
			excludedPhrases := combinePhraseFilters(query.ExcludePhrases, query.ExcludeInDescription)
			needDescription := requiredPhrases != "" || excludedPhrases != ""

			var detail CcDetail
			var detailFetchedAt *time.Time
			if needDescription {
				var fetched bool
				var available bool
				var attempted bool
				listingID := qry.StableListingId("cashConverters", sum.CanonicalUrl)
				detail, fetched, available, attempted, err = s.getCcDetail(ctx, sum, existingListings[listingID], descriptionFetchAvailable)
				if attempted {
					descriptionFetchAvailable = false
				}
				if err != nil {
					log.Printf("Failed to get product details for %s: %v", sum.CanonicalUrl, err)
					continue
				}
				if !available {
					continue
				}
				if fetched {
					fetchedAt := time.Now().UTC()
					detailFetchedAt = &fetchedAt
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
				LastDetailAt: detailFetchedAt,
			}
			listingId := qry.StableListingId("cashConverters", detail.CanonicalUrl)
			discovered = append(discovered, found)

			// Evaluate keyword and description filters
			matched := true
			var rejectReason *string

			searchable := strings.ToLower(detail.Title + " " + detail.Description)
			if requiredPhrases != "" {
				for _, phrase := range parsePhrases(requiredPhrases) {
					if !strings.Contains(searchable, phrase) {
						matched = false
						reason := "MissingRequiredPhrase"
						rejectReason = &reason
						break
					}
				}
			}

			if matched && excludedPhrases != "" {
				for _, phrase := range parsePhrases(excludedPhrases) {
					if strings.Contains(searchable, phrase) {
						matched = false
						reason := "ExcludedPhrase"
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

			prev := existingStates[listingId]

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

			states = append(states, state)
		}

		if err := s.dbClient.PersistListingBatch(ctx, discovered, existingListings, states, now); err != nil {
			log.Printf("Failed to persist Cash Converters results for URL %s: %v", *query.Url, err)
		}
	}

	return notifs, nil
}

func (s *CashConvertersScanner) discoverCcItems(ctx context.Context, searchUrl string, minPrice, maxPrice *float64) ([]CcSummary, error) {
	apiUrl := buildCcApiUrl(searchUrl, minPrice, maxPrice)

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
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

func (s *CashConvertersScanner) getCcDetail(
	ctx context.Context,
	sum CcSummary,
	listing *models.Listing,
	allowFetch bool,
) (CcDetail, bool, bool, bool, error) {
	// Item details are immutable for scanner purposes. Once fetched, always
	// reuse the persisted description instead of revisiting the item page.
	if listing != nil && listing.LastDetailAt != nil {
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
		}, false, true, false, nil
	}

	if !allowFetch {
		return CcDetail{}, false, false, false, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", sum.CanonicalUrl, nil)
	if err != nil {
		return CcDetail{}, false, false, true, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return CcDetail{}, false, false, true, err
	}
	if resp == nil {
		return CcDetail{}, false, false, true, fmt.Errorf("http response is nil")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CcDetail{}, false, false, true, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return CcDetail{}, false, false, true, err
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
	}, true, true, true, nil
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

func combinePhraseFilters(first, second *string) string {
	switch {
	case first == nil || *first == "":
		if second == nil {
			return ""
		}
		return *second
	case second == nil || *second == "":
		return *first
	default:
		return *first + "," + *second
	}
}
