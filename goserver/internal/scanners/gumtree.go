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
	"dealscanner/internal/models"
	"dealscanner/internal/notifications"
	"github.com/PuerkitoBio/goquery"
)

type GumtreeScanner struct {
	dbClient *qry.Query
}

func NewGumtreeScanner(dbClient *qry.Query) *GumtreeScanner {
	return &GumtreeScanner{dbClient: dbClient}
}

func (s *GumtreeScanner) Name() string {
	return "Gumtree"
}

func (s *GumtreeScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	globals, err := s.dbClient.GetGlobals(ctx)
	if err != nil {
		return nil, err
	}
	if globals == nil || !globals.Gumtree {
		return nil, nil
	}

	queries, err := s.dbClient.ListSavedQueries(ctx, "gumtree")
	if err != nil {
		return nil, err
	}

	var notifs []notifications.AppNotification
	now := time.Now().UTC()

	for _, item := range queries {
		if item.Url == nil {
			continue
		}

		time.Sleep(3 * time.Second) // rate limiting protection

		listings, err := s.scrapeGumtree(ctx, *item.Url)
		if err != nil {
			log.Printf("Gumtree scrape failed for URL %s: %v", *item.Url, err)
			continue
		}

		cache, err := s.dbClient.LoadListingEvaluationCache(ctx, item.QueryId, listings)
		if err != nil {
			return nil, err
		}
		states := make([]*models.QueryListingState, 0, len(listings))

		for _, found := range listings {
			listingId := qry.StableListingId("gumtree", found.CanonicalUrl)
			match := qry.MatchResult{Type: qry.MatchTypeMatched}
			if found.TotalPrice != nil {
				match = qry.MatchPriceRange(*found.TotalPrice, nil, item.MaxPrice)
			}

			decision, state := qry.BuildQueryListingStateDecision(
				cache.ExistingStates[listingId],
				item.QueryId,
				listingId,
				"gumtree",
				found.TotalPrice,
				match,
				now,
				0,
			)
			if state != nil {
				states = append(states, state)
			}

			if decision.Type == qry.DecisionTypeNotify && found.TotalPrice != nil {
				title := fmt.Sprintf("a %s priced at $%.2f is available at %s", found.Title, *found.TotalPrice, found.CanonicalUrl)
				notifs = append(notifs, notifications.AppNotification{
					Kind:     "deal",
					Source:   "gumtree",
					Title:    title,
					Url:      found.CanonicalUrl,
					Price:    found.TotalPrice,
					ImageUrl: found.ImageUrl,
					Query: &notifications.NotificationQuery{
						Type: "gumtree",
						Id:   item.Id,
					},
				})
			}
		}

		if err := s.dbClient.PersistListingBatch(ctx, listings, cache.ExistingListings, cache.LatestObservations, cache.ExistingStates, states, now); err != nil {
			return nil, err
		}
	}

	return notifs, nil
}

func (s *GumtreeScanner) scrapeGumtree(ctx context.Context, searchUrl string) ([]qry.DiscoveredListing, error) {
	doc, err := scrapeUrlWithBrowser(ctx, searchUrl)
	if err != nil {
		return nil, err
	}

	var listings []qry.DiscoveredListing

	// Look up cards
	doc.Find("a[class*='user-ad-row-new-design'], a[href*='/s-ad/']").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || href == "" {
			return
		}

		if !strings.HasPrefix(href, "http") {
			href = "https://www.gumtree.com.au" + href
		}

		title := CleanSelectionText(sel.Find("p.user-ad-row-new-design__title, [class*='title']").First())
		if title == "" {
			return
		}

		priceText := strings.TrimSpace(sel.Find("span.user-ad-price-new-design__price, [class*='price']").First().Text())
		if priceText == "" {
			return
		}

		imgSel := sel.Find("img").First()
		imgUrl, _ := imgSel.Attr("src")

		price, err := parseGumtreePrice(priceText)
		if err != nil {
			return
		}

		avail := "available"
		listings = append(listings, qry.DiscoveredListing{
			Source:       "gumtree",
			CanonicalUrl: href,
			Title:        title,
			Price:        &price,
			TotalPrice:   &price,
			ImageUrl:     &imgUrl,
			Availability: &avail,
		})
	})

	return listings, nil
}

func parseGumtreePrice(priceText string) (float64, error) {
	if strings.Contains(strings.ToLower(priceText), "free") {
		return 0, nil
	}
	reg := regexp.MustCompile(`[^0-9.]`)
	clean := reg.ReplaceAllString(priceText, "")
	pVal, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, err
	}
	return pVal, nil
}
