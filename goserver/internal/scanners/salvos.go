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

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/notifications"
	"github.com/PuerkitoBio/goquery"
)

type SalvosScanner struct {
	dbClient *qry.Query
}

func NewSalvosScanner(dbClient *qry.Query) *SalvosScanner {
	return &SalvosScanner{dbClient: dbClient}
}

func (s *SalvosScanner) Name() string {
	return "Salvos"
}

func (s *SalvosScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	globals, err := s.dbClient.GetGlobals(ctx)
	if err != nil {
		return nil, err
	}
	if globals == nil || !globals.Salvos {
		return nil, nil
	}

	queries, err := s.dbClient.ListSavedQueries(ctx, "salvos")
	if err != nil {
		return nil, err
	}

	var notifs []notifications.AppNotification
	now := time.Now().UTC()

	for _, item := range queries {
		if item.Name == nil {
			continue
		}

		time.Sleep(3 * time.Second) // rate limiting protection

		listings, err := s.scrapeSalvos(ctx, *item.Name)
		if err != nil {
			log.Printf("Salvos scrape failed for query %s: %v", *item.Name, err)
			continue
		}

		cache, err := s.dbClient.LoadListingEvaluationCache(ctx, item.QueryId, listings)
		if err != nil {
			return nil, err
		}
		states := make([]*models.QueryListingState, 0, len(listings))

		for _, found := range listings {
			listingId := qry.StableListingId("salvos", found.CanonicalUrl)
			match := qry.MatchResult{Type: qry.MatchTypeMatched}
			if found.TotalPrice != nil {
				match = qry.MatchPriceRange(*found.TotalPrice, item.MinPrice, item.MaxPrice)
			}

			decision, state := qry.BuildQueryListingStateDecision(
				cache.ExistingStates[listingId],
				item.QueryId,
				listingId,
				"salvos",
				found.TotalPrice,
				match,
				now,
				0,
			)
			if state != nil {
				states = append(states, state)
			}

			if decision.Type == qry.DecisionTypeNotify && found.TotalPrice != nil {
				title := fmt.Sprintf("a %s is available for $%.2f at %s", found.Title, *found.TotalPrice, found.CanonicalUrl)
				notifs = append(notifs, notifications.AppNotification{
					Kind:     "deal",
					Source:   "salvos",
					Title:    title,
					Url:      found.CanonicalUrl,
					Price:    found.TotalPrice,
					ImageUrl: found.ImageUrl,
					Query: &notifications.NotificationQuery{
						Type: "salvos",
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

func (s *SalvosScanner) scrapeSalvos(ctx context.Context, term string) ([]qry.DiscoveredListing, error) {
	searchUrl := fmt.Sprintf("https://www.salvosstores.com.au/shop?search=%s&sorting=newestFirst&price=0-99999", url.QueryEscape(term))

	doc, err := scrapeUrlWithBrowser(ctx, searchUrl)
	if err != nil {
		return nil, err
	}

	var listings []qry.DiscoveredListing

	// Look up cards
	doc.Find("div.flex.flex-col.overflow-hidden.rounded.shadow-card.bg-white.h-auto, [class*='rounded'][class*='shadow-card']").Each(func(i int, sel *goquery.Selection) {
		linkSel := sel.Find("a.line-clamp-3, [class*='line-clamp-3']").First()
		title := CleanSelectionText(linkSel)
		if title == "" {
			linkSel = sel.Find("a[href]").FilterFunction(func(i int, s *goquery.Selection) bool {
				return strings.TrimSpace(s.Text()) != ""
			}).First()
			title = CleanSelectionText(linkSel)
		}
		href, exists := linkSel.Attr("href")
		if !exists || title == "" {
			return
		}

		if !strings.HasPrefix(href, "http") {
			href = "https://www.salvosstores.com.au" + href
		}

		priceText := strings.TrimSpace(sel.Find("div.product-price, [class*='product-price']").First().Text())
		if priceText == "" {
			return
		}

		imgSel := sel.Find("img").First()
		imgUrl, _ := imgSel.Attr("src")

		price, err := parseSalvosPrice(priceText)
		if err != nil {
			return
		}

		avail := "available"
		listings = append(listings, qry.DiscoveredListing{
			Source:       "salvos",
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

func parseSalvosPrice(priceText string) (float64, error) {
	reg := regexp.MustCompile(`[^0-9.]`)
	clean := reg.ReplaceAllString(priceText, "")
	return strconv.ParseFloat(clean, 64)
}
