package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	qry "dealscanner/internal/db/query"
)

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
			ProductListItemCount int                `json:"ProductListItemCount"`
			ProductListItems     []CcApiProductItem `json:"ProductListItems"`
		} `json:"ProductList"`
	} `json:"Value"`
}

type ccCatalogSort string

const (
	ccCatalogSortNewest ccCatalogSort = "newest"
	ccCatalogSortPrice  ccCatalogSort = "price"
	ccCatalogPageSize                 = 24
)

type ccCatalogPage struct {
	Summaries  []qry.CashConvertersDiscoveredSummary
	TotalItems int
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

func valueOr(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
}

func phrasesMatch(searchable string, phrases []string, mode string) bool {
	if len(phrases) == 0 {
		return false
	}
	combine, wholeWords := phraseMode(mode)
	matches := func(phrase string) bool {
		if wholeWords {
			return containsWholePhrase(searchable, phrase)
		}
		return strings.Contains(searchable, phrase)
	}
	if combine == "any" {
		for _, phrase := range phrases {
			if matches(phrase) {
				return true
			}
		}
		return false
	}
	for _, phrase := range phrases {
		if !matches(phrase) {
			return false
		}
	}
	return true
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

func buildCcCatalogPageURL(sort ccCatalogSort, page int) string {
	return buildCcSearchPageURL(sort, page, "")
}

func buildCcSearchPageURL(sort ccCatalogSort, page int, query string) string {
	u := url.URL{
		Scheme: "https",
		Host:   "www.cashconverters.com.au",
		Path:   "/c3api/search/results",
	}
	q := url.Values{}
	q.Set("Sort", string(sort))
	q.Set("page", strconv.Itoa(page))
	q.Set("query", query)
	u.RawQuery = q.Encode()
	return u.String()
}

func normalizeCcCanonicalURL(rawUrl string) string {
	rawUrl = strings.TrimSpace(rawUrl)
	if rawUrl == "" {
		return ""
	}
	if !strings.HasPrefix(rawUrl, "http://") && !strings.HasPrefix(rawUrl, "https://") {
		if !strings.HasPrefix(rawUrl, "/") {
			rawUrl = "/" + rawUrl
		}
		rawUrl = "https://www.cashconverters.com.au" + rawUrl
	}

	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return rawUrl
	}

	parsed.Scheme = "https"
	parsed.Host = "www.cashconverters.com.au"
	parsed.Fragment = ""

	// Remove tracking / extra parameters while preserving essential item path
	q := parsed.Query()
	for k := range q {
		if strings.HasPrefix(k, "utm_") || k == "ref" || k == "source" {
			q.Del(k)
		}
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

var ccItemIDRegex = regexp.MustCompile(`/([0-9]{6,})($|[/?#-])`)

func extractCcExternalID(item CcApiProductItem, canonicalURL string) *string {
	// 1. Check URL path or slug for numeric ID
	if matches := ccItemIDRegex.FindStringSubmatch(canonicalURL); len(matches) > 1 {
		id := matches[1]
		return &id
	}
	if item.Url != "" {
		if matches := ccItemIDRegex.FindStringSubmatch(item.Url); len(matches) > 1 {
			id := matches[1]
			return &id
		}
	}
	return nil
}

func (s *CashConvertersScanner) fetchCatalogPage(
	ctx context.Context,
	sort ccCatalogSort,
	page int,
) (ccCatalogPage, error) {
	return fetchCcSearchPage(ctx, s.client, s.baseURL, "", sort, page)
}

func fetchCcSearchPage(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	query string,
	sort ccCatalogSort,
	page int,
) (ccCatalogPage, error) {
	apiUrl := buildCcSearchPageURL(sort, page, query)
	if baseURL != "" {
		parsed, err := url.Parse(apiUrl)
		if err == nil {
			baseParsed, errBase := url.Parse(baseURL)
			if errBase == nil {
				parsed.Scheme = baseParsed.Scheme
				parsed.Host = baseParsed.Host
				apiUrl = parsed.String()
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return ccCatalogPage{}, fmt.Errorf("create request for %s page %d: %w", sort, page, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return ccCatalogPage{}, fmt.Errorf("fetch %s page %d: %w", sort, page, err)
	}
	if resp == nil {
		return ccCatalogPage{}, fmt.Errorf("http response is nil for %s page %d", sort, page)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ccCatalogPage{}, fmt.Errorf("bad status code %d for %s page %d", resp.StatusCode, sort, page)
	}

	var apiResp CcApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return ccCatalogPage{}, fmt.Errorf("decode JSON %s page %d: %w", sort, page, err)
	}

	if !apiResp.WasSuccessful {
		return ccCatalogPage{}, fmt.Errorf("API response indicates failure for %s page %d", sort, page)
	}

	var results []qry.CashConvertersDiscoveredSummary
	for _, item := range apiResp.Value.ProductList.ProductListItems {
		price := parseCcPrice(item.Sp)
		shipping := parseCcShipping(item.ShippingCost)
		total := price + shipping
		canonical := normalizeCcCanonicalURL(item.Url)
		if canonical == "" {
			continue
		}
		extID := extractCcExternalID(item, canonical)

		results = append(results, qry.CashConvertersDiscoveredSummary{
			CanonicalURL:   canonical,
			ExternalItemID: extID,
			Title:          strings.Join(strings.Fields(item.Title), " "),
			Price:          price,
			Shipping:       shipping,
			TotalPrice:     total,
			ImageURL:       item.AbsoluteImageUrl,
		})
	}

	return ccCatalogPage{
		Summaries:  results,
		TotalItems: apiResp.Value.ProductList.ProductListItemCount,
	}, nil
}

// FetchCashConvertersSearchPage queries the live Cash Converters search index.
// TotalItems is the API's authoritative result count and lets callers determine
// the exact final page without guessing from a short or empty response.
func FetchCashConvertersSearchPage(
	ctx context.Context,
	client *http.Client,
	query string,
	sort string,
	page int,
) ([]qry.CashConvertersDiscoveredSummary, int, error) {
	apiSort := ccCatalogSort(sort)
	if apiSort == "" {
		apiSort = ccCatalogSortNewest
	}
	result, err := fetchCcSearchPage(ctx, client, "", query, apiSort, page)
	if err != nil {
		return nil, 0, err
	}
	return result.Summaries, result.TotalItems, nil
}
