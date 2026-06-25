package scanners

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/notifications"
	"github.com/PuerkitoBio/goquery"
)

type EbayScanner struct {
	dbClient *qry.Query
}

func NewEbayScanner(dbClient *qry.Query) *EbayScanner {
	return &EbayScanner{dbClient: dbClient}
}

func (e *EbayScanner) Name() string {
	return "Ebay"
}

func (e *EbayScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	queries, err := e.dbClient.ListSavedQueries(ctx, "ebay")
	if err != nil {
		return nil, err
	}

	var notifs []notifications.AppNotification

	for _, item := range queries {
		if item.Url == nil {
			continue
		}

		listings, err := e.scrapeEbay(ctx, *item.Url)
		if err != nil {
			log.Printf("eBay scrape failed for URL %s: %v", *item.Url, err)
			continue
		}

		for _, found := range listings {
			listingId := qry.StableListingId("ebay", found.CanonicalUrl)
			obsId, err := e.dbClient.PersistListingObservation(ctx, found, time.Now().UTC())
			if err != nil {
				log.Printf("Failed to persist eBay observation: %v", err)
				continue
			}

			now := time.Now().UTC()
			match := qry.MatchResult{Type: qry.MatchTypeMatched}
			if found.TotalPrice != nil {
				match = qry.MatchPriceRange(*found.TotalPrice, nil, item.MaxPrice)
			}

			decision, err := e.dbClient.EvaluateListingForQuery(ctx, item.QueryId, listingId, "ebay", found.TotalPrice, match, now, 0)
			if err != nil {
				return nil, err
			}

			if decision.Type == qry.DecisionTypeNotify && found.TotalPrice != nil {
				title := fmt.Sprintf("a %s priced at $%.2f is available at %s", found.Title, *found.TotalPrice, found.CanonicalUrl)
				notifs = append(notifs, notifications.AppNotification{
					Kind:     "deal",
					Source:   "ebay",
					Title:    title,
					Url:      found.CanonicalUrl,
					Price:    found.TotalPrice,
					ImageUrl: found.ImageUrl,
					Query: &notifications.NotificationQuery{
						Type: "ebay",
						Id:   item.Id,
					},
				})
			}

			_ = obsId
		}

		if !sleepWithContext(ctx, 3*time.Second) {
			return nil, ctx.Err()
		}
	}

	return notifs, nil
}

func sleepWithContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (e *EbayScanner) scrapeEbay(ctx context.Context, url string) ([]qry.DiscoveredListing, error) {
	doc, err := scrapeUrlWithBrowser(ctx, url)
	if err != nil {
		return nil, err
	}

	var listings []qry.DiscoveredListing

	// Select card elements (supports traditional s-item and newer s-card layouts)
	doc.Find("li.s-item, li.s-card").Each(func(i int, s *goquery.Selection) {
		titleSel := s.Find("div.s-item__title, div[role=\"heading\"]").Clone()
		// Remove noisy/extra screen reader text and badge elements
		titleSel.Find("span.clipped, span.s-card__new-listing, span.s-item__new-listing, span.s-item__watch-heart, .LIGHT_HIGHLIGHT").Remove()
		title := CleanSelectionText(titleSel)
		if title == "" || strings.HasPrefix(title, "Shop on eBay") {
			return
		}

		priceText := strings.TrimSpace(s.Find(".s-item__price, .s-card__price").First().Text())
		if priceText == "" {
			return
		}

		linkSel := s.Find("a.s-item__link, a.s-card__link").First()
		if linkSel.Length() == 0 {
			linkSel = s.Find("a[href]").First()
		}
		link, exists := linkSel.Attr("href")
		if !exists || link == "" {
			return
		}

		imgSel := s.Find("img").First()
		img, _ := imgSel.Attr("src")

		parsedPrice, currency := parsePrice(priceText)
		if parsedPrice == 0 {
			return
		}

		extId := parseExternalId(link)
		avail := "available"

		listings = append(listings, qry.DiscoveredListing{
			Source:       "ebay",
			ExternalId:   extId,
			CanonicalUrl: link,
			Title:        title,
			Price:        &parsedPrice,
			TotalPrice:   &parsedPrice,
			Currency:     &currency,
			ImageUrl:     &img,
			Availability: &avail,
		})
	})

	return listings, nil
}

func parsePrice(priceText string) (float64, string) {
	// e.g. "AU $12.34" or "$12.34 to $15.00"
	if idx := strings.Index(priceText, " to "); idx != -1 {
		priceText = priceText[:idx]
	}

	reg := regexp.MustCompile(`[^0-9.]`)
	clean := reg.ReplaceAllString(priceText, "")
	val, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, ""
	}

	if strings.Contains(priceText, "AU") {
		return val, "AUD"
	}
	// Default conversion for USD
	return val * 1.55, "USD"
}

func parseExternalId(url string) *string {
	reg := regexp.MustCompile(`\/itm\/(\d+)`)
	matches := reg.FindStringSubmatch(url)
	if len(matches) > 1 {
		return &matches[1]
	}
	return nil
}
