package scanners

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/playwright-community/playwright-go"
	"golang.org/x/net/html"
)

// CleanSelectionText extracts text from a goquery.Selection and normalizes spacing between elements and words.
func CleanSelectionText(s *goquery.Selection) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			txt := strings.TrimSpace(n.Data)
			if txt != "" {
				parts = append(parts, txt)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	for _, n := range s.Nodes {
		walk(n)
	}
	raw := strings.Join(parts, " ")
	return strings.Join(strings.Fields(raw), " ")
}

// GetPageHTMLWithBrowser launches the best suited browser engine (headless or headed)
// depending on the target URL to bypass blocks, navigates to the URL, and returns the outer HTML.
func GetPageHTMLWithBrowser(ctx context.Context, targetUrl string, timeout time.Duration) (string, error) {
	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("failed to start playwright: %w", err)
	}
	defer pw.Stop()

	var browser playwright.Browser
	var userAgent string

	if strings.Contains(targetUrl, "ebay.com.au") || strings.Contains(targetUrl, "ebay.com") {
		// eBay blocks Chromium but works flawlessly headlessly with WebKit (Safari engine)
		browser, err = pw.WebKit.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
		})
		userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15"
	} else if strings.Contains(targetUrl, "gumtree.com.au") || strings.Contains(targetUrl, "gumtree.com") {
		// Gumtree requires headed Chromium to bypass anti-bot blocks
		// We launch it off-screen so it is completely invisible to the user
		browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false),
			Args: []string{
				"--disable-blink-features=AutomationControlled",
				"--no-sandbox",
				"--window-position=-10000,-10000",
			},
		})
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
	} else {
		// Default to Chromium in headless mode (e.g. for Salvos)
		browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
			Args: []string{
				"--disable-blink-features=AutomationControlled",
				"--no-sandbox",
			},
		})
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
	}

	if err != nil {
		return "", fmt.Errorf("failed to launch browser for %s: %w", targetUrl, err)
	}
	defer browser.Close()

	browserCtx, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(userAgent),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		Locale: playwright.String("en-US"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create browser context: %w", err)
	}
	defer browserCtx.Close()

	page, err := browserCtx.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to open new page: %w", err)
	}

	_, err = page.Goto(targetUrl, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
		Timeout:   playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return "", fmt.Errorf("failed to navigate to URL %s: %w", targetUrl, err)
	}

	// Sleep 3 seconds to let dynamic content render
	time.Sleep(3 * time.Second)

	htmlContent, err := page.Content()
	if err != nil {
		return "", fmt.Errorf("failed to get page content: %w", err)
	}

	return htmlContent, nil
}

func scrapeUrlWithBrowser(ctx context.Context, url string) (*goquery.Document, error) {
	htmlStr, err := GetPageHTMLWithBrowser(ctx, url, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
}

// CheckBrowser checks if Playwright driver can be initialized and Chromium launched.
// It automatically installs/verifies browser binaries when run.
func CheckBrowser(ctx context.Context) error {
	if err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true}); err != nil {
		return fmt.Errorf("failed to install playwright driver: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to start playwright runtime: %w", err)
	}
	defer pw.Stop()

	// Launch chromium in headless mode for checks
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to launch chromium: %w", err)
	}
	browser.Close()
	return nil
}
