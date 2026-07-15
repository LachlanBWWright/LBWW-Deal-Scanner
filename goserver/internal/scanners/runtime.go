package scanners

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"dealscanner/internal/config"
	qry "dealscanner/internal/db/query"
	"dealscanner/internal/notifications"
	"dealscanner/internal/runtime"
)

type StatusSetter interface {
	SetStatus(statusText string)
}

const statusUpdateInterval = 20 * time.Second
const minimumScanCycleDuration = 30 * time.Second
const DefaultScanTimeout = 3 * time.Minute
const runtimeStateFlushIntervalCycles = 10

type Runner struct {
	cfg          *config.Config
	dbClient     *qry.Query
	stateManager *runtime.StateManager
	notifService *notifications.NotificationService
	statusSetter StatusSetter
	scanners     []Scanner
	scanTimeout  time.Duration
	statusMu     sync.Mutex
	previousScan *scanStatusResult
}

type scanTimeoutError struct {
	scanner string
	timeout time.Duration
}

func (e scanTimeoutError) Error() string {
	return fmt.Sprintf("scanner %q exceeded the %s time limit", e.scanner, e.timeout)
}

type scanStatusResult struct {
	name    string
	elapsed time.Duration
	outcome string
}

type Scanner interface {
	Name() string
	Scan(ctx context.Context) ([]notifications.AppNotification, error)
}

func NewRunner(
	cfg *config.Config,
	dbClient *qry.Query,
	stateManager *runtime.StateManager,
	notifService *notifications.NotificationService,
	statusSetter StatusSetter,
) *Runner {
	return &Runner{
		cfg:          cfg,
		dbClient:     dbClient,
		stateManager: stateManager,
		notifService: notifService,
		statusSetter: statusSetter,
		scanTimeout:  DefaultScanTimeout,
	}
}

func (r *Runner) RegisterScanner(s Scanner) {
	r.scanners = append(r.scanners, s)
}

func (r *Runner) Start(ctx context.Context) {
	if r.cfg.ScheduledMode {
		go r.StartScheduledScan(ctx, time.Duration(r.cfg.ScheduledDurationMs)*time.Millisecond)
	} else {
		go r.StartBackgroundScanLoop(ctx)
	}
}

func (r *Runner) StartBackgroundScanLoop(ctx context.Context) {
	r.stateManager.BeginScanLoop()
	defer r.stateManager.EndScanLoop()

	log.Println("Background scanner loop started.")
	r.runLoop(ctx, nil)
}

func (r *Runner) StartScheduledScan(ctx context.Context, duration time.Duration) {
	r.stateManager.BeginScanLoop()
	defer r.stateManager.EndScanLoop()

	log.Printf("Scheduled scanner started for duration: %v", duration)

	// Record start in db
	dbState, err := r.dbClient.GetScannerRuntimeState(ctx)
	if err == nil {
		now := time.Now().UTC()
		dbState.LastStartedAt = &now
		r.dbClient.UpdateScannerRuntimeState(ctx, dbState)
	}

	start := time.Now()
	r.runLoop(ctx, &duration)
	elapsed := time.Since(start)

	// Record stop in db
	dbState, err = r.dbClient.GetScannerRuntimeState(ctx)
	if err == nil {
		now := time.Now().UTC()
		dbState.LastStoppedAt = &now
		dbState.ScheduledRuns++
		dbState.TotalScheduledRuntimeMs += elapsed.Milliseconds()
		r.dbClient.UpdateScannerRuntimeState(ctx, dbState)
	}
}

func (r *Runner) runLoop(ctx context.Context, maxDuration *time.Duration) {
	loopStart := time.Now()

	dbState, err := r.dbClient.GetScannerRuntimeState(ctx)
	if err != nil {
		log.Printf("Failed to load scanner runtime state: %v", err)
		return
	}

	steamScanCnt := dbState.SteamScanCount
	csTradeScanCnt := dbState.CsTradeScanCount
	cyclesSinceRuntimeStateFlush := 0

	for {
		if maxDuration != nil && time.Since(loopStart) >= *maxDuration {
			break
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Println("Starting scan pass cycle...")
		cycleStartedAt := time.Now()
		record := r.stateManager.BeginScanRun("loop")

		r.runScanPass(ctx, steamScanCnt, csTradeScanCnt)

		errStr := (*string)(nil)
		r.stateManager.FinishScanRun(record, errStr)

		// Increment scan counters
		steamScanCnt++
		steamCounterReset := false
		if steamScanCnt >= 55 {
			steamScanCnt = 0
			steamCounterReset = true
		}
		csTradeScanCnt++
		csTradeCounterReset := false
		if csTradeScanCnt >= 100 {
			csTradeScanCnt = 0
			csTradeCounterReset = true
		}

		cyclesSinceRuntimeStateFlush++
		if cyclesSinceRuntimeStateFlush >= runtimeStateFlushIntervalCycles || steamCounterReset || csTradeCounterReset {
			dbState.SteamScanCount = steamScanCnt
			dbState.CsTradeScanCount = csTradeScanCnt
			r.dbClient.UpdateScannerRuntimeState(ctx, dbState)
			cyclesSinceRuntimeStateFlush = 0
		}

		if !waitForMinimumCycleDuration(ctx, cycleStartedAt, minimumScanCycleDuration) {
			return
		}
	}
}

func waitForMinimumCycleDuration(ctx context.Context, startedAt time.Time, minimum time.Duration) bool {
	remaining := minimum - time.Since(startedAt)
	if remaining <= 0 {
		return true
	}

	timer := time.NewTimer(remaining)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (r *Runner) runScanPass(ctx context.Context, steamScanCnt, csTradeScanCnt int) {
	// Execute registered scanners sequentially
	for _, sc := range r.scanners {
		if sc.Name() == "Steam Market" && steamScanCnt < 55 {
			continue
		}
		if (sc.Name() == "CS Trade" || sc.Name() == "Loot Farm" || sc.Name() == "Trade It") && csTradeScanCnt < 100 {
			continue
		}

		log.Printf("Executing scanner: %s", sc.Name())
		notifs, err := r.scanWithTimedStatus(ctx, sc)
		if err != nil {
			log.Printf("Scanner %q encountered an error: %v", sc.Name(), err)
			continue
		}

		if len(notifs) > 0 {
			log.Printf("Scanner %q found %d new deals", sc.Name(), len(notifs))
			r.publish(ctx, notifs)
		}
	}
}

func (r *Runner) publish(ctx context.Context, notifs []notifications.AppNotification) {
	// Record in StateManager recent results
	var records []runtime.SearchResultRecord
	nowStr := time.Now().UTC().Format(time.RFC3339)
	for _, n := range notifs {
		var qType, qId *string
		if n.Query != nil {
			qType = &n.Query.Type
			qId = &n.Query.Id
		}

		records = append(records, runtime.SearchResultRecord{
			Source:    n.Source,
			Title:     n.Title,
			Url:       n.Url,
			Price:     n.Price,
			ImageUrl:  n.ImageUrl,
			QueryType: qType,
			QueryId:   qId,
			FoundAt:   nowStr,
		})
	}
	r.stateManager.RecordSearchResults(records)

	// Publish to notification service
	for _, n := range notifs {
		r.notifService.Publish(ctx, n)
	}
}

func (r *Runner) RunManualScanOnce(ctx context.Context) {
	record := r.stateManager.BeginScanRun("manual")
	log.Println("Manual scan run started...")

	for _, sc := range r.scanners {
		log.Printf("[Manual] Executing scanner: %s", sc.Name())
		notifs, err := r.scanWithTimedStatus(ctx, sc)
		if err != nil {
			log.Printf("[Manual] Scanner %q failed: %v", sc.Name(), err)
			continue
		}
		if len(notifs) > 0 {
			r.publish(ctx, notifs)
		}
	}

	r.stateManager.FinishScanRun(record, nil)
	log.Println("Manual scan run completed.")
}

func (r *Runner) scanWithTimedStatus(
	ctx context.Context,
	sc Scanner,
) ([]notifications.AppNotification, error) {
	startedAt := time.Now()
	previous := r.getPreviousScanResult()

	var stopStatus func()
	if r.statusSetter != nil {
		statusCtx, cancelStatus := context.WithCancel(ctx)
		statusStopped := make(chan struct{})
		r.statusSetter.SetStatus(formatScanningStatus(sc.Name(), 0, false, previous))
		go func() {
			defer close(statusStopped)
			ticker := time.NewTicker(statusUpdateInterval)
			defer ticker.Stop()

			for {
				select {
				case <-statusCtx.Done():
					return
				case <-ticker.C:
					r.statusSetter.SetStatus(
						formatScanningStatus(sc.Name(), time.Since(startedAt), true, previous),
					)
				}
			}
		}()
		stopStatus = func() {
			cancelStatus()
			<-statusStopped
		}
	} else {
		stopStatus = func() {}
	}

	notifs, err := r.runScannerWithTimeout(ctx, sc)
	stopStatus()
	r.setPreviousScanResult(scanStatusResult{
		name:    sc.Name(),
		elapsed: time.Since(startedAt),
		outcome: getScanOutcome(notifs, err),
	})
	return notifs, err
}

type scanResult struct {
	notifications []notifications.AppNotification
	err           error
}

func (r *Runner) runScannerWithTimeout(
	ctx context.Context,
	sc Scanner,
) ([]notifications.AppNotification, error) {
	scanCtx, cancelScan := context.WithTimeout(ctx, r.scanTimeout)
	defer cancelScan()

	result := make(chan scanResult, 1)
	go func() {
		notifs, err := sc.Scan(scanCtx)
		result <- scanResult{notifications: notifs, err: err}
	}()

	select {
	case completed := <-result:
		return completed.notifications, completed.err
	case <-scanCtx.Done():
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		timeoutErr := scanTimeoutError{
			scanner: sc.Name(),
			timeout: r.scanTimeout,
		}
		r.notifService.Publish(context.WithoutCancel(ctx), notifications.AppNotification{
			Kind:    "error",
			Source:  sc.Name(),
			Message: "Warning: " + timeoutErr.Error() + ". The scan was cancelled.",
			Tags:    []string{"warning", "scan-timeout"},
		})
		return nil, timeoutErr
	}
}

func (r *Runner) getPreviousScanResult() *scanStatusResult {
	r.statusMu.Lock()
	defer r.statusMu.Unlock()
	if r.previousScan == nil {
		return nil
	}
	result := *r.previousScan
	return &result
}

func (r *Runner) setPreviousScanResult(result scanStatusResult) {
	r.statusMu.Lock()
	defer r.statusMu.Unlock()
	r.previousScan = &result
}

func getScanOutcome(notifs []notifications.AppNotification, scanErr error) string {
	if scanErr != nil {
		return "FAILED"
	}
	if len(notifs) > 0 {
		return "NEW ITEM FOUND"
	}
	return "NOT FOUND"
}

func formatScanningStatus(
	name string,
	elapsed time.Duration,
	includeElapsed bool,
	previous *scanStatusResult,
) string {
	current := getStatusText(name)
	if includeElapsed {
		current = formatTimedStatus(name, elapsed)
	}
	if previous == nil {
		return current
	}
	return fmt.Sprintf(
		"%s | %s (%s): %s",
		current,
		previous.name,
		formatElapsed(previous.elapsed),
		previous.outcome,
	)
}

func formatTimedStatus(name string, elapsed time.Duration) string {
	return fmt.Sprintf("%s (%s)", getStatusText(name), formatElapsed(elapsed))
}

func formatElapsed(elapsed time.Duration) string {
	totalSeconds := int(elapsed / time.Second)
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func getStatusText(name string) string {
	switch name {
	case "Steam Market":
		return "Scanning the Steam Community Market"
	case "CS Trade":
		return "Scanning CS.Trade"
	case "Loot Farm":
		return "Scanning loot.farm"
	case "Trade It":
		return "Scanning tradeit.gg"
	default:
		return "Scanning " + name
	}
}

// Helper to sleep with random jitter
func JitterSleep(minMs, maxMs int) {
	delta := maxMs - minMs
	if delta <= 0 {
		time.Sleep(time.Duration(minMs) * time.Millisecond)
		return
	}
	jitter := rand.Intn(delta)
	time.Sleep(time.Duration(minMs+jitter) * time.Millisecond)
}
