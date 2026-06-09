package scanners

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// GetPageHTMLWithBrowser launches a headless browser, navigates to the URL,
// waits for the page to render, and returns the outer HTML.
func GetPageHTMLWithBrowser(ctx context.Context, url string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	chromeCtx, chromeCancel := chromedp.NewContext(allocCtx)
	defer chromeCancel()

	var htmlContent string

	err := chromedp.Run(chromeCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return "", fmt.Errorf("failed to scrape URL %s with browser emulator: %w", url, err)
	}

	return htmlContent, nil
}

func scrapeUrlWithBrowser(ctx context.Context, url string) (*goquery.Document, error) {
	htmlStr, err := GetPageHTMLWithBrowser(ctx, url, 30*time.Second)
	if err != nil {
		return nil, err
	}
	// Import strings in browser.go
	return goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
}
