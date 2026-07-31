package scanners

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/notifications"
)

const (
	ccTailPagesPerPass        = 10
	ccMaxDetailFetchesPerPass = 6
	ccMaxMissingChecksPerPass = 3
	ccRequestMinDelayMs       = 500
	ccRequestMaxDelayMs       = 1500
)

type ccDiscoveryPage struct {
	sort ccCatalogSort
	page int
}

type CashConvertersScanner struct {
	dbClient *qry.Query
	client   *http.Client
	mu       sync.Mutex
	baseURL  string
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

func (s *CashConvertersScanner) Scan(ctx context.Context) ([]notifications.AppNotification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	for page := 1; ; page++ {
		catalogPage, fetchErr := s.fetchCatalogPage(ctx, ccCatalogSortNewest, page)
		if fetchErr != nil {
			errStr := fetchErr.Error()
			state.LastError = &errStr
			state.LastErrorAt = &now
			log.Printf("Cash Converters newest scan error on page %d: %v", page, fetchErr)
			break
		}
		state.LastSuccessfulRequestAt = &now
		state.LastError = nil
		upserted, upsertErr := s.dbClient.UpsertCashConvertersSummaries(ctx, catalogPage.Summaries, state.CurrentSweepID, page, now)
		if upsertErr != nil {
			return nil, upsertErr
		}
		for _, id := range upserted.ListingIDs {
			if enqueueErr := s.dbClient.EnqueueCashConvertersDetailJobs(ctx, id, now); enqueueErr != nil {
				return nil, enqueueErr
			}
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
		JitterSleep(ccRequestMinDelayMs, ccRequestMaxDelayMs)
	}

	// Continue the stable price-ordered back-catalog sweep independently.
	pages = s.planDiscoveryPages(state)
	for _, discovery := range pages {
		catalogPage, err := s.fetchCatalogPage(ctx, discovery.sort, discovery.page)
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

		upserted, err := s.dbClient.UpsertCashConvertersSummaries(
			ctx,
			catalogPage.Summaries,
			state.CurrentSweepID,
			discovery.page,
			now,
		)
		if err != nil {
			return nil, err
		}

		for _, id := range upserted.ListingIDs {
			if err := s.dbClient.EnqueueCashConvertersDetailJobs(ctx, id, now); err != nil {
				return nil, err
			}
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
			break
		}

		JitterSleep(ccRequestMinDelayMs, ccRequestMaxDelayMs)
	}

	// 2. Process due detail jobs
	dueJobs, err := s.dbClient.ListDueCashConvertersDetailJobs(ctx, ccMaxDetailFetchesPerPass, now)
	if err != nil {
		return nil, err
	}
	if len(dueJobs) > 0 {
		for _, job := range dueJobs {
			listingID, err := s.processDetailJob(ctx, job, now)
			if err != nil {
				log.Printf("Cash Converters detail job failed for listing %s: %v", listingID, err)
			} else {
				changed[listingID] = struct{}{}
			}
			JitterSleep(ccRequestMinDelayMs, ccRequestMaxDelayMs)
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
		for _, listing := range missingVerifications {
			if err := s.verifyMissingListing(ctx, listing, now); err != nil {
				log.Printf("Cash Converters missing verification failed for listing %s: %v", listing.ID, err)
			} else {
				changed[listing.ID] = struct{}{}
			}
			JitterSleep(ccRequestMinDelayMs, ccRequestMaxDelayMs)
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

	notifs, err := s.evaluateChangedListings(ctx, filters, changedIDs, now)
	if err != nil {
		return nil, err
	}

	return notifs, nil
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
