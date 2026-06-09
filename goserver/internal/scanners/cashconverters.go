package scanners

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dealscanner/internal/models"
	qry "dealscanner/internal/db/query"
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
			log.Printf("Cash Converters scrape failed for URL %s: %v", *query.Url, err)
			continue
		}

		for _, sum := range summaries {
			// Get or fetch detail
			detail, err := s.getCcDetail(ctx, sum)
			if err != nil {
				continue
			}

			// Persist
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

			// Evaluate keywords and max price
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

			prev, err := s.dbClient.GetQueryListingState(ctx, query.Id, listingId)
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
				QueryId:            query.Id,
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
	parsed, err := url.Parse(searchUrl)
	if err != nil {
		return nil, err
	}
	q := parsed.Query()
	q.Set("page", "1")
	parsed.RawQuery = q.Encode()

	doc, err := scrapeUrlWithBrowser(ctx, parsed.String())
	if err != nil {
		return nil, err
	}

	var results []CcSummary

	doc.Find("div.product-item, .product-item").Each(func(i int, sel *goquery.Selection) {
		titleSel := sel.Find("span.product-item__title__description, [class*='title']").First()
		title := strings.TrimSpace(titleSel.Text())

		linkSel := sel.Find("a").First()
		href, exists := linkSel.Attr("href")
		if !exists || title == "" {
			return
		}

		if !strings.HasPrefix(href, "http") {
			href = "https://www.cashconverters.com.au" + href
		}

		priceText := strings.TrimSpace(sel.Find(".product-item__price").First().Text())
		shippingText := strings.TrimSpace(sel.Find(".product-item__postage").First().Text())
		imgSel := sel.Find("img").First()
		imgUrl, _ := imgSel.Attr("src")

		price := parseCcPrice(priceText)
		shipping := parseCcShipping(shippingText)
		total := price + shipping

		results = append(results, CcSummary{
			CanonicalUrl: href,
			Title:        title,
			Price:        price,
			Shipping:     shipping,
			TotalPrice:   total,
			ImageUrl:     imgUrl,
		})
	})

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

	doc, err := scrapeUrlWithBrowser(ctx, sum.CanonicalUrl)
	if err != nil {
		return CcDetail{}, err
	}

	title := strings.TrimSpace(doc.Find("h1, .product-detail__title, .product-title").First().Text())
	if title == "" {
		title = sum.Title
	}

	description := strings.TrimSpace(doc.Find(".product-detail__description, .product-description, [class*='description']").First().Text())
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
