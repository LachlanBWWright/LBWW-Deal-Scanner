package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dealscanner/internal/notifications"
	"github.com/PuerkitoBio/goquery"
)

type TemporaryScanResult struct {
	Items                  []ManualScanItemResult          `json:"items"`
	Notifications          []notifications.AppNotification `json:"notifications"`
	Errors                 []string                        `json:"errors"`
	NotificationsPublished bool                            `json:"notificationsPublished"`
}

type ManualScanItemResult struct {
	Source        string   `json:"source"`
	Title         string   `json:"title"`
	Url           string   `json:"url"`
	Price         *float64 `json:"price"`
	ImageUrl      *string  `json:"imageUrl"`
	PassedFilters bool     `json:"passedFilters"`
	FilterReason  *string  `json:"filterReason"`
}

func optionalNumber(val interface{}, fallback float64) float64 {
	if val == nil {
		return fallback
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
	}
	return fallback
}

func requiredNumber(val interface{}, name string, errors *[]string) (float64, bool) {
	if val == nil {
		*errors = append(*errors, fmt.Sprintf("%s is required", name))
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f, true
		}
	}
	*errors = append(*errors, fmt.Sprintf("%s must be a number", name))
	return 0, false
}

func phraseList(val interface{}) []string {
	if val == nil {
		return nil
	}
	str, ok := val.(string)
	if !ok {
		return nil
	}
	var res []string
	for _, p := range strings.Split(str, ",") {
		clean := strings.TrimSpace(strings.ToLower(p))
		if clean != "" {
			res = append(res, clean)
		}
	}
	return res
}

func matchesPhraseFilters(text string, requiredPhrases, excludePhrases interface{}) bool {
	normalized := strings.ToLower(text)
	req := phraseList(requiredPhrases)
	excl := phraseList(excludePhrases)

	for _, p := range req {
		if !strings.Contains(normalized, p) {
			return false
		}
	}
	for _, p := range excl {
		if strings.Contains(normalized, p) {
			return false
		}
	}
	return true
}

func RunTemporaryScan(ctx context.Context, qType string, payload map[string]interface{}) TemporaryScanResult {
	var items []ManualScanItemResult
	var notifs []notifications.AppNotification
	var errors []string

	switch qType {
	case "ebay":
		urlStr, _ := payload["url"].(string)
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" {
			errors = append(errors, "ebay scan requires payload.url")
			break
		}
		maxPrice := optionalNumber(payload["maxPrice"], 999999999)

		scraped, err := collectEbayItems(ctx, urlStr, maxPrice)
		if err != nil {
			errors = append(errors, fmt.Sprintf("eBay scan failed: %v", err))
		} else {
			items = scraped
			for _, item := range items {
				if item.PassedFilters {
					notifs = append(notifs, notificationFromItem(item, qType))
				}
			}
		}

	case "gumtree":
		urlStr, _ := payload["url"].(string)
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" {
			errors = append(errors, "gumtree scan requires payload.url")
			break
		}
		maxPrice := optionalNumber(payload["maxPrice"], 999999999)

		scraped, err := collectGumtreeItems(ctx, urlStr, maxPrice)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Gumtree scan failed: %v", err))
		} else {
			items = scraped
			for _, item := range items {
				if item.PassedFilters {
					notifs = append(notifs, notificationFromItem(item, qType))
				}
			}
		}

	case "cashConverters":
		urlStr, _ := payload["url"].(string)
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" {
			errors = append(errors, "cashConverters scan requires payload.url")
			break
		}

		scraped, err := collectCashConvertersItems(ctx, urlStr, payload["requiredPhrases"], payload["excludePhrases"])
		if err != nil {
			errors = append(errors, fmt.Sprintf("Cash Converters scan failed: %v", err))
		} else {
			items = scraped
			for _, item := range items {
				if item.PassedFilters {
					notifs = append(notifs, notificationFromItem(item, qType))
				}
			}
		}

	case "salvos":
		name, _ := payload["name"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			errors = append(errors, "salvos scan requires payload.name")
			break
		}
		minPrice := optionalNumber(payload["minPrice"], 0)
		maxPrice := optionalNumber(payload["maxPrice"], 999999999)

		scraped, err := collectSalvosItems(ctx, name, minPrice, maxPrice)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Salvos scan failed: %v", err))
		} else {
			items = scraped
			for _, item := range items {
				if item.PassedFilters {
					notifs = append(notifs, notificationFromItem(item, qType))
				}
			}
		}

	case "steamMarket":
		name, _ := payload["name"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			errors = append(errors, "steamMarket scan requires payload.name")
			break
		}
		maxPrice, ok := requiredNumber(payload["maxPrice"], "maxPrice", &errors)
		if !ok {
			break
		}
		displayUrl, _ := payload["displayUrl"].(string)
		if displayUrl == "" {
			displayUrl = name
		}

		scraped, err := collectSteamMarketItems(ctx, name, maxPrice, displayUrl)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Steam Market scan failed: %v", err))
		} else {
			items = scraped
			for _, item := range items {
				if item.PassedFilters {
					notifs = append(notifs, notificationFromItem(item, qType))
				}
			}
		}

	case "csMarket":
		urlStr, _ := payload["url"].(string)
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" {
			errors = append(errors, "csMarket scan requires payload.url")
			break
		}
		maxPrice, ok1 := requiredNumber(payload["maxPrice"], "maxPrice", &errors)
		maxFloat, ok2 := requiredNumber(payload["maxFloat"], "maxFloat", &errors)
		if !ok1 || !ok2 {
			break
		}
		displayUrl, _ := payload["displayUrl"].(string)
		if displayUrl == "" {
			displayUrl = urlStr
		}

		scraped, err := collectCsMarketItems(ctx, urlStr, maxPrice, maxFloat, displayUrl)
		if err != nil {
			errors = append(errors, fmt.Sprintf("CS Market scan failed: %v", err))
		} else {
			items = scraped
			for _, item := range items {
				if item.PassedFilters {
					notifs = append(notifs, notificationFromItem(item, qType))
				}
			}
		}

	case "csTradeBot":
		name, _ := payload["name"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			errors = append(errors, "csTradeBot scan requires payload.name")
			break
		}
		maxPrice, ok1 := requiredNumber(payload["maxPrice"], "maxPrice", &errors)
		minFloat, ok2 := requiredNumber(payload["minFloat"], "minFloat", &errors)
		maxFloat, ok3 := requiredNumber(payload["maxFloat"], "maxFloat", &errors)
		if !ok1 || !ok2 || !ok3 {
			break
		}

		scraped, err := collectCsTradeBotItems(ctx, name, maxPrice, minFloat, maxFloat)
		if err != nil {
			errors = append(errors, fmt.Sprintf("CS Trade Bot scan failed: %v", err))
		} else {
			items = scraped
			for _, item := range items {
				if item.PassedFilters {
					notifs = append(notifs, notifications.AppNotification{
						Kind:     "deal",
						Source:   "csTrade",
						Title:    item.Title,
						Url:      item.Url,
						Price:    item.Price,
						ImageUrl: item.ImageUrl,
						Query: &notifications.NotificationQuery{
							Type: qType,
							Id:   name,
						},
					})
				}
			}
		}

	default:
		errors = append(errors, fmt.Sprintf("Scanner type %q is not supported for temporary scans", qType))
	}

	return TemporaryScanResult{
		Items:                  items,
		Notifications:          notifs,
		Errors:                 errors,
		NotificationsPublished: false,
	}
}

func filterReason(passed bool, reason string) *string {
	if passed {
		return nil
	}
	return &reason
}

func notificationFromItem(item ManualScanItemResult, qType string) notifications.AppNotification {
	return notifications.AppNotification{
		Kind:     "deal",
		Source:   item.Source,
		Title:    item.Title,
		Url:      item.Url,
		Price:    item.Price,
		ImageUrl: item.ImageUrl,
		Query: &notifications.NotificationQuery{
			Type: qType,
			Id:   item.Url,
		},
	}
}

func collectEbayItems(ctx context.Context, searchUrl string, maxPrice float64) ([]ManualScanItemResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", searchUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []ManualScanItemResult

	doc.Find("li.s-item").Each(func(i int, s *goquery.Selection) {
		if len(results) >= 20 {
			return
		}
		titleSel := s.Find("div.s-item__title, div[role=\"heading\"]").First()
		title := strings.TrimSpace(titleSel.Text())
		if title == "" || strings.HasPrefix(title, "Shop on eBay") {
			return
		}
		title = strings.TrimPrefix(title, "New listing")
		title = strings.TrimSpace(title)

		priceText := strings.TrimSpace(s.Find(".s-item__price").First().Text())
		if priceText == "" {
			return
		}

		linkSel := s.Find("a.s-item__link, a[href]").First()
		link, exists := linkSel.Attr("href")
		if !exists || link == "" {
			return
		}

		imgSel := s.Find("img").First()
		img, _ := imgSel.Attr("src")

		parsedPrice, _ := parsePrice(priceText)

		passed := parsedPrice > 0 && parsedPrice <= maxPrice
		reason := ""
		if parsedPrice <= 0 {
			reason = "Price could not be read"
		} else if parsedPrice > maxPrice {
			reason = fmt.Sprintf("Price %.2f is above max %.2f", parsedPrice, maxPrice)
		}

		results = append(results, ManualScanItemResult{
			Source:        "ebay",
			Title:         title,
			Url:           link,
			Price:         &parsedPrice,
			ImageUrl:      &img,
			PassedFilters: passed,
			FilterReason:  filterReason(passed, reason),
		})
	})

	return results, nil
}

func collectGumtreeItems(ctx context.Context, searchUrl string, maxPrice float64) ([]ManualScanItemResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", searchUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []ManualScanItemResult

	doc.Find("a[class*='user-ad-row-new-design'], a[href*='/s-ad/']").Each(func(i int, sel *goquery.Selection) {
		if len(results) >= 20 {
			return
		}
		href, exists := sel.Attr("href")
		if !exists || href == "" {
			return
		}
		if !strings.HasPrefix(href, "http") {
			href = "https://www.gumtree.com.au" + href
		}

		title := strings.TrimSpace(sel.Find("p.user-ad-row-new-design__title, [class*='title']").First().Text())
		if title == "" {
			return
		}

		priceText := strings.TrimSpace(sel.Find("span.user-ad-price-new-design__price, [class*='price']").First().Text())
		if priceText == "" {
			return
		}

		imgSel := sel.Find("img").First()
		imgUrl, _ := imgSel.Attr("src")

		var price float64
		if strings.Contains(strings.ToLower(priceText), "free") {
			price = 0
		} else {
			reg := regexp.MustCompile(`[^0-9.]`)
			clean := reg.ReplaceAllString(priceText, "")
			pVal, err := strconv.ParseFloat(clean, 64)
			if err != nil {
				return
			}
			price = pVal
		}

		passed := price <= maxPrice
		reason := ""
		if price > maxPrice {
			reason = fmt.Sprintf("Price %.2f is above max %.2f", price, maxPrice)
		}

		results = append(results, ManualScanItemResult{
			Source:        "gumtree",
			Title:         title,
			Url:           href,
			Price:         &price,
			ImageUrl:      &imgUrl,
			PassedFilters: passed,
			FilterReason:  filterReason(passed, reason),
		})
	})

	return results, nil
}

func collectCashConvertersItems(ctx context.Context, searchUrl string, requiredPhrases, excludePhrases interface{}) ([]ManualScanItemResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", searchUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []ManualScanItemResult

	doc.Find("div.product-item, .product-item").Each(func(i int, sel *goquery.Selection) {
		if len(results) >= 20 {
			return
		}
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

		passed := matchesPhraseFilters(title, requiredPhrases, excludePhrases)

		results = append(results, ManualScanItemResult{
			Source:        "cashConverters",
			Title:         title,
			Url:           href,
			Price:         &total,
			ImageUrl:      &imgUrl,
			PassedFilters: passed,
			FilterReason:  filterReason(passed, "Phrase filters did not match"),
		})
	})

	return results, nil
}

func collectSalvosItems(ctx context.Context, name string, minPrice, maxPrice float64) ([]ManualScanItemResult, error) {
	searchUrl := fmt.Sprintf("https://www.salvosstores.com.au/shop?search=%s&sorting=newestFirst&price=0-99999", url.QueryEscape(name))

	req, err := http.NewRequestWithContext(ctx, "GET", searchUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []ManualScanItemResult

	doc.Find("div.flex.flex-col.overflow-hidden.rounded.shadow-card.bg-white.h-auto, [class*='rounded'][class*='shadow-card']").Each(func(i int, sel *goquery.Selection) {
		if len(results) >= 20 {
			return
		}
		linkSel := sel.Find("a[class*='line-clamp-3'], a[href]").First()
		title := strings.TrimSpace(linkSel.Text())
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

		reg := regexp.MustCompile(`[^0-9.]`)
		clean := reg.ReplaceAllString(priceText, "")
		price, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			return
		}

		passed := price >= minPrice && price <= maxPrice
		reason := ""
		if price < minPrice || price > maxPrice {
			reason = fmt.Sprintf("Price %.2f is outside %.2f - %.2f", price, minPrice, maxPrice)
		}

		results = append(results, ManualScanItemResult{
			Source:        "salvos",
			Title:         title,
			Url:           href,
			Price:         &price,
			ImageUrl:      &imgUrl,
			PassedFilters: passed,
			FilterReason:  filterReason(passed, reason),
		})
	})

	return results, nil
}

func collectSteamMarketItems(ctx context.Context, name string, maxPrice float64, displayUrl string) ([]ManualScanItemResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", name, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	var data SteamMarketResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []ManualScanItemResult
	for _, res := range data.Results {
		if len(results) >= 20 {
			break
		}
		price := res.SellPrice / 100.0
		passed := price <= maxPrice
		reason := ""
		if price > maxPrice {
			reason = fmt.Sprintf("Price %.2f is above max %.2f", price, maxPrice)
		}
		results = append(results, ManualScanItemResult{
			Source:        "steamMarket",
			Title:         res.Name,
			Url:           displayUrl,
			Price:         &price,
			ImageUrl:      nil,
			PassedFilters: passed,
			FilterReason:  filterReason(passed, reason),
		})
	}

	return results, nil
}

func collectCsMarketItems(ctx context.Context, itemUrl string, maxPrice, maxFloat float64, displayUrl string) ([]ManualScanItemResult, error) {
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam returned status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw struct {
		ListingInfo json.RawMessage `json:"listinginfo"`
	}
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, err
	}

	var listingInfo map[string]SteamListing
	if len(raw.ListingInfo) > 0 && raw.ListingInfo[0] == '{' {
		if err := json.Unmarshal(raw.ListingInfo, &listingInfo); err != nil {
			return nil, err
		}
	}

	var results []ManualScanItemResult
	i := 0
	for _, listing := range listingInfo {
		if i >= 10 {
			break
		}
		if len(listing.Asset.MarketActions) == 0 {
			continue
		}
		action := listing.Asset.MarketActions[0]
		query := fmt.Sprintf("https://api.csgofloat.com/?url=%s", action.Link)
		query = strings.ReplaceAll(query, "%listingid%", listing.ListingID)
		query = strings.ReplaceAll(query, "%assetid%", listing.Asset.ID)

		price := (listing.ConvertedPricePerUnit + listing.ConvertedFeePerUnit) / 100.0

		time.Sleep(500 * time.Millisecond) // sleep to avoid csgofloat ratelimit

		// Fetch float from CSGO float API
		infoReq, err := http.NewRequestWithContext(ctx, "GET", query, nil)
		if err != nil {
			continue
		}
		infoReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

		infoResp, err := client.Do(infoReq)
		if err != nil {
			continue
		}

		var itemInfo SteamItemInfoResponse
		decodeErr := json.NewDecoder(infoResp.Body).Decode(&itemInfo)
		infoResp.Body.Close()
		if decodeErr != nil {
			continue
		}

		passed := itemInfo.ItemInfo.FloatValue <= maxFloat && price <= maxPrice
		reason := ""
		if itemInfo.ItemInfo.FloatValue > maxFloat {
			reason = fmt.Sprintf("Float %.5f is above max %.5f", itemInfo.ItemInfo.FloatValue, maxFloat)
		} else if price > maxPrice {
			reason = fmt.Sprintf("Price %.2f is above max %.2f", price, maxPrice)
		}

		results = append(results, ManualScanItemResult{
			Source:        "steamMarket",
			Title:         fmt.Sprintf("%s with float %f", itemInfo.ItemInfo.FullItemName, itemInfo.ItemInfo.FloatValue),
			Url:           displayUrl,
			Price:         &price,
			ImageUrl:      nil,
			PassedFilters: passed,
			FilterReason:  filterReason(passed, reason),
		})
		i++
	}

	return results, nil
}

func collectCsTradeBotItems(ctx context.Context, name string, maxPrice, minFloat, maxFloat float64) ([]ManualScanItemResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://cdn.cs.trade:8443/api/getInventory?order_by=price_desc&bot=all&_=1651756783463", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data CsTradeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []ManualScanItemResult
	for _, item := range data.Inventory {
		if item.AppID != 730 {
			continue
		}
		nameMatches := item.MarketHashName == name
		if !nameMatches {
			continue
		}

		passed := item.Price <= maxPrice && item.Wear >= minFloat && item.Wear <= maxFloat
		reason := ""
		if item.Price > maxPrice {
			reason = fmt.Sprintf("Price %.2f is above max %.2f", item.Price, maxPrice)
		} else if item.Wear < minFloat || item.Wear > maxFloat {
			reason = fmt.Sprintf("Float %.5f is outside %.5f - %.5f", item.Wear, minFloat, maxFloat)
		}

		icon := item.Icon
		results = append(results, ManualScanItemResult{
			Source:        "csTrade",
			Title:         fmt.Sprintf("%s with float %f", item.MarketHashName, item.Wear),
			Url:           "https://cs.trade/",
			Price:         &item.Price,
			ImageUrl:      &icon,
			PassedFilters: passed,
			FilterReason:  filterReason(passed, reason),
		})
	}

	return results, nil
}
