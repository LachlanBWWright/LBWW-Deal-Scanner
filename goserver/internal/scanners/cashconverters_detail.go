package scanners

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"github.com/PuerkitoBio/goquery"
)

type CcDetailPageResult struct {
	CanonicalUrl string
	Title        string
	Description  string
	Availability string
	ImageUrl     string
	Price        float64
	Shipping     float64
	TotalPrice   float64
	StatusCode   int
	IsMissing    bool
}

func (s *CashConvertersScanner) fetchCcDetailPage(ctx context.Context, canonicalUrl string) (*CcDetailPageResult, error) {
	targetUrl := canonicalUrl
	if s.baseURL != "" {
		parsed, err := url.Parse(canonicalUrl)
		if err == nil {
			baseParsed, errBase := url.Parse(s.baseURL)
			if errBase == nil {
				parsed.Scheme = baseParsed.Scheme
				parsed.Host = baseParsed.Host
				targetUrl = parsed.String()
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for %s: %w", canonicalUrl, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request for %s: %w", canonicalUrl, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("http response is nil for %s", canonicalUrl)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return &CcDetailPageResult{
			CanonicalUrl: canonicalUrl,
			Availability: "unavailable",
			StatusCode:   resp.StatusCode,
			IsMissing:    true,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code %d for %s", resp.StatusCode, canonicalUrl)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse HTML for %s: %w", canonicalUrl, err)
	}

	title := CleanSelectionText(doc.Find("h1, .product-detail__title, .product-title").First())

	var description string
	doc.Find("meta[name='description']").Each(func(i int, sel *goquery.Selection) {
		if content, exists := sel.Attr("content"); exists {
			description = strings.TrimSpace(content)
		}
	})
	if description == "" {
		doc.Find("meta[property='og:description']").Each(func(i int, sel *goquery.Selection) {
			if content, exists := sel.Attr("content"); exists {
				description = strings.TrimSpace(content)
			}
		})
	}
	if description == "" {
		description = strings.TrimSpace(doc.Find(".product-detail__description, .product-description, [class*='description']").First().Text())
	}

	regSpace := regexp.MustCompile(`\s+`)
	description = regSpace.ReplaceAllString(description, " ")

	bodyText := strings.ToLower(doc.Find("body").Text())
	avail := "available"
	if strings.Contains(bodyText, "sold") || strings.Contains(bodyText, "out of stock") || strings.Contains(bodyText, "no longer available") {
		avail = "unavailable"
	}

	var imgUrl string
	doc.Find("meta[property='og:image']").Each(func(i int, sel *goquery.Selection) {
		if content, exists := sel.Attr("content"); exists {
			imgUrl = strings.TrimSpace(content)
		}
	})

	return &CcDetailPageResult{
		CanonicalUrl: canonicalUrl,
		Title:        title,
		Description:  description,
		Availability: avail,
		ImageUrl:     imgUrl,
		StatusCode:   resp.StatusCode,
		IsMissing:    avail == "unavailable",
	}, nil
}

func (s *CashConvertersScanner) processDetailJob(ctx context.Context, job models.CashConvertersDetailJob, now time.Time) (string, error) {
	var lst models.Listing
	db := s.dbClient.UnderlyingDB().WithContext(ctx)
	if err := db.Where("id = ?", job.ListingID).First(&lst).Error; err != nil {
		if completeErr := s.dbClient.CompleteCashConvertersDetailJob(ctx, job.ID, now); completeErr != nil {
			return job.ListingID, fmt.Errorf("complete orphaned detail job %s: %w", job.ID, completeErr)
		}
		return job.ListingID, nil
	}

	res, err := s.fetchCcDetailPage(ctx, lst.CanonicalUrl)
	if err != nil {
		nextDue := now.Add(time.Duration((job.Attempts+1)*10) * time.Minute)
		if retryErr := s.dbClient.RetryCashConvertersDetailJob(ctx, job.ID, nextDue, err.Error()); retryErr != nil {
			return job.ListingID, fmt.Errorf("retry detail job %s after fetch failure: %w", job.ID, retryErr)
		}
		return job.ListingID, err
	}

	if res.IsMissing {
		detailInput := qry.CcDetailInput{
			CanonicalUrl: lst.CanonicalUrl,
			Title:        lst.Title,
			Description:  valueOr(lst.Description, ""),
			Availability: "unavailable",
		}
		if err := s.dbClient.UpdateCashConvertersDetail(ctx, job.ListingID, detailInput, now); err != nil {
			return job.ListingID, err
		}
		if err := s.dbClient.CompleteCashConvertersDetailJob(ctx, job.ID, now); err != nil {
			return job.ListingID, err
		}
		return job.ListingID, nil
	}

	title := res.Title
	if title == "" {
		title = lst.Title
	}
	img := res.ImageUrl
	if img == "" && lst.ImageUrl != nil {
		img = *lst.ImageUrl
	}

	detailInput := qry.CcDetailInput{
		CanonicalUrl: lst.CanonicalUrl,
		Title:        title,
		Description:  res.Description,
		Availability: res.Availability,
		ImageUrl:     img,
	}

	if err := s.dbClient.UpdateCashConvertersDetail(ctx, job.ListingID, detailInput, now); err != nil {
		if retryErr := s.dbClient.RetryCashConvertersDetailJob(ctx, job.ID, now.Add(10*time.Minute), err.Error()); retryErr != nil {
			return job.ListingID, fmt.Errorf("retry detail job %s after update failure: %w", job.ID, retryErr)
		}
		return job.ListingID, err
	}

	if err := s.dbClient.CompleteCashConvertersDetailJob(ctx, job.ID, now); err != nil {
		return job.ListingID, err
	}
	return job.ListingID, nil
}

func (s *CashConvertersScanner) verifyMissingListing(ctx context.Context, listing models.Listing, now time.Time) error {
	res, err := s.fetchCcDetailPage(ctx, listing.CanonicalUrl)
	if err != nil {
		return fmt.Errorf("verify missing Cash Converters listing %s: %w", listing.ID, err)
	}
	if res == nil {
		return fmt.Errorf("verify missing Cash Converters listing %s: empty detail result", listing.ID)
	}
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusGone {
		return s.dbClient.ConfirmCashConvertersMissing(ctx, listing.ID, now)
	}
	if res.IsMissing {
		_, err := s.dbClient.RecordCashConvertersMissingVerification(ctx, listing.ID, now)
		return err
	}

	// Item came back 200 OK and available!
	detailInput := qry.CcDetailInput{
		CanonicalUrl: listing.CanonicalUrl,
		Title:        res.Title,
		Description:  res.Description,
		Availability: "available",
		ImageUrl:     res.ImageUrl,
	}
	return s.dbClient.UpdateCashConvertersDetail(ctx, listing.ID, detailInput, now)
}
