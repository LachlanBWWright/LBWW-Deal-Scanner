package scanners

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/notifications"
)

const (
	ccHeadPagesPerPass = 5
	ccTailPagesPerPass = 10
	// Detail work is deliberately bounded. A large durable backlog must drain
	// over many cycles and never turn one scan into an unbounded batch.
	ccMaxDetailFetchesPerPass = 8
	ccDetailWorkers           = 3
	ccDetailJobTimeout        = 20 * time.Second
	ccMaxMissingChecksPerPass = 3
	ccRequestMinDelayMs       = 500
	ccRequestMaxDelayMs       = 1500
	ccHeadDetailPriority      = 100
	ccTailDetailPriority      = 10
)

type ccDiscoveryPage struct {
	sort ccCatalogSort
	page int
}

type detailJobResult struct {
	job         models.CashConvertersDetailJob
	listingID   string
	description string
	elapsed     time.Duration
	err         error
}

type CashConvertersScanner struct {
	dbClient *qry.Query
	client   *http.Client
	mu       sync.Mutex
	baseURL  string
	insights []string
}

func NewCashConvertersScanner(dbClient *qry.Query) *CashConvertersScanner {
	return &CashConvertersScanner{
		dbClient: dbClient,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *CashConvertersScanner) getBaseURL() string {
	if s.baseURL != "" {
		return s.baseURL
	}
	return "https://www.cashconverters.com.au"
}

func (s *CashConvertersScanner) Name() string {
	return "Cash Converters"
}

func (s *CashConvertersScanner) ScanInsights() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.insights...)
}

func (s *CashConvertersScanner) recordCatalogPage(
	sort ccCatalogSort,
	page int,
	catalogPage ccCatalogPage,
	elapsed time.Duration,
) {
	pageURL := buildCcCatalogPageURL(sort, page)
	if s.baseURL != "" {
		pageURL = s.getBaseURL() + "/c3api/search/results?Sort=" + string(sort) + "&page=" + fmt.Sprint(page) + "&query="
	}
	s.insights = append(s.insights, fmt.Sprintf(
		"\n**Catalog %s page %d** — %d items returned (%d reported total), request took %s\n%s",
		sort,
		page,
		len(catalogPage.Summaries),
		catalogPage.TotalItems,
		elapsed.Round(time.Millisecond),
		pageURL,
	))
	for _, item := range catalogPage.Summaries {
		s.insights = append(s.insights, fmt.Sprintf(
			"• %s — $%.2f item + $%.2f shipping ($%.2f total)\n%s",
			item.Title,
			item.Price,
			item.Shipping,
			item.TotalPrice,
			item.CanonicalURL,
		))
	}
}

func (s *CashConvertersScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.insights = nil
	scanStartedAt := time.Now()
	catalogRequests := 0
	detailRequests := 0
	missingRequests := 0
	catalogRequestTime := time.Duration(0)
	detailRequestTime := time.Duration(0)
	missingRequestTime := time.Duration(0)
	defer func() {
		s.insights = append(s.insights, fmt.Sprintf(
			"\n**Timing breakdown**\nCatalogue page requests: %d in %s\nDetail jobs processed (normally one item-page request each): %d in %s\nMissing-item verification page requests: %d in %s\nTotal Cash Converters scan time: %s\nIntentional rate-limit delay: 0.5–1.5s between catalogue/detail requests",
			catalogRequests,
			catalogRequestTime.Round(time.Millisecond),
			detailRequests,
			detailRequestTime.Round(time.Millisecond),
			missingRequests,
			missingRequestTime.Round(time.Millisecond),
			time.Since(scanStartedAt).Round(time.Millisecond),
		))
	}()

	globals, err := s.dbClient.GetGlobals(ctx)
	if err != nil {
		return nil, err
	}
	if globals == nil || !globals.CashConverters {
		return nil, nil
	}

	filters, err := s.dbClient.ListSavedQueries(ctx, "cashConverters")
	if err != nil {
		return nil, err
	}

	state, err := s.dbClient.GetCashConvertersScanState(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	changed := make(map[string]struct{})
	justCompletedSweep := false

	// 1. Always start at the newest inventory and continue until a page reaches
	// inventory already known before this pass.
	pages := make([]ccDiscoveryPage, 0, ccTailPagesPerPass+1)
	for page := 1; page <= ccHeadPagesPerPass; page++ {
		requestStartedAt := time.Now()
		catalogRequests++
		catalogPage, fetchErr := s.fetchCatalogPage(ctx, ccCatalogSortNewest, page)
		requestElapsed := time.Since(requestStartedAt)
		catalogRequestTime += requestElapsed
		if fetchErr != nil {
			errStr := fetchErr.Error()
			state.LastError = &errStr
			state.LastErrorAt = &now
			log.Printf("Cash Converters newest scan error on page %d: %v", page, fetchErr)
			break
		}
		state.LastSuccessfulRequestAt = &now
		state.LastError = nil
		s.recordCatalogPage(ccCatalogSortNewest, page, catalogPage, requestElapsed)
		upserted, upsertErr := s.dbClient.UpsertCashConvertersSummariesWithDetailJobs(ctx, catalogPage.Summaries, state.CurrentSweepID, page, now, ccHeadDetailPriority)
		if upsertErr != nil {
			return nil, upsertErr
		}
		for _, id := range upserted.ChangedListingIDs {
			changed[id] = struct{}{}
		}
		state.LastHeadScanAt = &now
		lastPage := (catalogPage.TotalItems + ccCatalogPageSize - 1) / ccCatalogPageSize
		if len(catalogPage.Summaries) == 0 || upserted.HadPreviouslySeen ||
			(catalogPage.TotalItems > 0 && page >= lastPage) {
			break
		}
		if !sleepWithContext(ctx, ccRequestMinDelayMs*time.Millisecond) {
			break
		}
	}

	// Continue the stable price-ordered back-catalog sweep independently.
	pages = s.planDiscoveryPages(state)
	for _, discovery := range pages {
		requestStartedAt := time.Now()
		catalogRequests++
		catalogPage, err := s.fetchCatalogPage(ctx, discovery.sort, discovery.page)
		requestElapsed := time.Since(requestStartedAt)
		catalogRequestTime += requestElapsed
		if err != nil {
			errStr := err.Error()
			state.LastError = &errStr
			state.LastErrorAt = &now
			log.Printf(
				"Cash Converters scan error on %s page %d: %v",
				discovery.sort,
				discovery.page,
				err,
			)
			break // Do not advance cursor for failed page
		}

		state.LastSuccessfulRequestAt = &now
		state.LastError = nil
		s.recordCatalogPage(discovery.sort, discovery.page, catalogPage, requestElapsed)

		upserted, err := s.dbClient.UpsertCashConvertersSummariesWithDetailJobs(
			ctx,
			catalogPage.Summaries,
			state.CurrentSweepID,
			discovery.page,
			now,
			ccTailDetailPriority,
		)
		if err != nil {
			return nil, err
		}

		for _, id := range upserted.ChangedListingIDs {
			changed[id] = struct{}{}
		}

		// Update page cursor state
		state.LastTailScanAt = &now
		if len(catalogPage.Summaries) > 0 {
			state.LastKnownNonEmptyPage = discovery.page
			state.ConsecutiveEmptyTailPages = 0
			state.TailNextPage = discovery.page + 1
		} else {
			state.ConsecutiveEmptyTailPages++
			state.TailNextPage = discovery.page + 1
		}

		lastPage := (catalogPage.TotalItems + ccCatalogPageSize - 1) / ccCatalogPageSize
		reachedReportedEnd := catalogPage.TotalItems > 0 && discovery.page >= lastPage
		reachedEmptyEnd := len(catalogPage.Summaries) == 0
		if reachedReportedEnd || reachedEmptyEnd {
			state.LastSweepCompletedAt = &now
			state.CurrentSweepID++
			state.TailNextPage = 1
			state.ConsecutiveEmptyTailPages = 0
			justCompletedSweep = true
		}

		// Persist discovery progress before detail work. Detail requests can consume
		// the remainder of the scanner deadline; without this checkpoint a timeout
		// makes the next pass repeat these same catalogue pages indefinitely.
		if err := s.dbClient.UpdateCashConvertersScanState(ctx, state); err != nil {
			return nil, err
		}
		if justCompletedSweep {
			break
		}

		if !sleepWithContext(ctx, ccRequestMinDelayMs*time.Millisecond) {
			break
		}
	}

	// 2. Process due detail jobs
	dueJobs, err := s.dbClient.ListDueCashConvertersDetailJobs(ctx, ccMaxDetailFetchesPerPass, now)
	if err != nil {
		return nil, err
	}
	if len(dueJobs) > 0 {
		s.insights = append(s.insights, fmt.Sprintf("\n**Detail pages processed:** %d", len(dueJobs)))
		results := s.processDetailJobs(ctx, dueJobs, now)
		for result := range results {
			detailRequests++
			detailRequestTime += result.elapsed
			if result.err != nil {
				log.Printf("Cash Converters detail job failed for listing %s: %v", result.listingID, result.err)
				s.insights = append(s.insights, fmt.Sprintf("• Detail failed in %s: %s — %v", result.elapsed.Round(time.Millisecond), result.job.ListingID, result.err))
			} else if result.listingID != "" {
				changed[result.listingID] = struct{}{}
				s.insights = append(s.insights, fmt.Sprintf("• Detail loaded in %s: %s\n  Description: %s", result.elapsed.Round(time.Millisecond), result.listingID, result.description))
			} else {
				s.insights = append(s.insights, fmt.Sprintf("• Orphaned detail job cleaned in %s: %s", result.elapsed.Round(time.Millisecond), result.job.ListingID))
			}
		}
	}

	// 3. Handle missing items after completed sweep
	if justCompletedSweep {
		if err := s.dbClient.MarkCashConvertersMissingCandidates(ctx, state.CurrentSweepID-1, now); err != nil {
			log.Printf("Cash Converters missing mark failed: %v", err)
		}
	}

	missingVerifications, err := s.dbClient.ListCashConvertersMissingVerificationsDue(ctx, ccMaxMissingChecksPerPass, now)
	if err != nil {
		return nil, err
	}
	if len(missingVerifications) > 0 {
		s.insights = append(s.insights, fmt.Sprintf("**Missing listings verified:** %d", len(missingVerifications)))
		for _, listing := range missingVerifications {
			requestStartedAt := time.Now()
			missingRequests++
			if err := s.verifyMissingListing(ctx, listing, now); err != nil {
				requestElapsed := time.Since(requestStartedAt)
				missingRequestTime += requestElapsed
				log.Printf("Cash Converters missing verification failed for listing %s: %v", listing.ID, err)
				s.insights = append(s.insights, fmt.Sprintf("• Verification failed in %s: %s — %v", requestElapsed.Round(time.Millisecond), listing.CanonicalUrl, err))
			} else {
				requestElapsed := time.Since(requestStartedAt)
				missingRequestTime += requestElapsed
				changed[listing.ID] = struct{}{}
				s.insights = append(s.insights, fmt.Sprintf("• Verified in %s: %s", requestElapsed.Round(time.Millisecond), listing.CanonicalUrl))
			}
			if !JitterSleepContext(ctx, ccRequestMinDelayMs, ccRequestMaxDelayMs) {
				break
			}
		}
	}

	// 4. Save updated scan state
	if err := s.dbClient.UpdateCashConvertersScanState(ctx, state); err != nil {
		return nil, err
	}

	// 5. Periodic retention prune of confirmed missing listings (older than 7 days)
	if _, err := s.dbClient.DeleteConfirmedCashConvertersListings(ctx, now.AddDate(0, 0, -7), 50); err != nil {
		return nil, err
	}

	// 6. Evaluate changed listings against saved filters
	changedIDs := make([]string, 0, len(changed))
	for id := range changed {
		changedIDs = append(changedIDs, id)
	}
	s.insights = append(s.insights, fmt.Sprintf("**New or changed listings evaluated:** %d", len(changedIDs)))

	notifs, err := s.evaluateChangedListings(ctx, filters, changedIDs, now)
	if err != nil {
		return nil, err
	}

	return notifs, nil
}

// processDetailJobs uses a small, fixed worker pool. Network fetches can
// overlap, while database writes remain bounded by the database connection
// pool. The result channel is closed only after every worker exits, so a
// cancelled scan cannot leak goroutines.
func (s *CashConvertersScanner) processDetailJobs(
	ctx context.Context,
	jobs []models.CashConvertersDetailJob,
	now time.Time,
) <-chan detailJobResult {
	results := make(chan detailJobResult, len(jobs))
	work := make(chan models.CashConvertersDetailJob)
	workerCount := ccDetailWorkers
	if len(jobs) < workerCount {
		workerCount = len(jobs)
	}

	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for job := range work {
				started := time.Now()
				jobCtx, cancel := context.WithTimeout(ctx, ccDetailJobTimeout)
				listingID, description, err := s.processDetailJob(jobCtx, job, now)
				cancel()
				results <- detailJobResult{
					job:         job,
					listingID:   listingID,
					description: description,
					elapsed:     time.Since(started),
					err:         err,
				}
				if !sleepWithContext(ctx, time.Duration(ccRequestMinDelayMs)*time.Millisecond) {
					return
				}
			}
		}()
	}

	go func() {
		defer close(work)
		for _, job := range jobs {
			select {
			case work <- job:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		workers.Wait()
		close(results)
	}()
	return results
}

func (s *CashConvertersScanner) planDiscoveryPages(state *models.CashConvertersScanState) []ccDiscoveryPage {
	pages := make([]ccDiscoveryPage, 0, ccTailPagesPerPass)
	startTail := state.TailNextPage
	if startTail < 1 {
		startTail = 1
	}
	for p := startTail; p < startTail+ccTailPagesPerPass; p++ {
		pages = append(pages, ccDiscoveryPage{
			sort: ccCatalogSortPrice,
			page: p,
		})
	}

	return pages
}
