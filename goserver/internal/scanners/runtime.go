package scanners

import (
	"context"
	"log"
	"math/rand"
	"time"

	"dealscanner/internal/config"
	"dealscanner/internal/db"
	"dealscanner/internal/notifications"
	"dealscanner/internal/runtime"
)

type Runner struct {
	cfg          *config.Config
	dbClient     *db.DB
	stateManager *runtime.StateManager
	notifService *notifications.NotificationService
	scanners     []Scanner
}

type Scanner interface {
	Name() string
	Scan(ctx context.Context) ([]notifications.AppNotification, error)
}

func NewRunner(
	cfg *config.Config,
	dbClient *db.DB,
	stateManager *runtime.StateManager,
	notifService *notifications.NotificationService,
) *Runner {
	return &Runner{
		cfg:          cfg,
		dbClient:     dbClient,
		stateManager: stateManager,
		notifService: notifService,
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
		record := r.stateManager.BeginScanRun("loop")

		r.runScanPass(ctx, steamScanCnt, csTradeScanCnt)

		errStr := (*string)(nil)
		r.stateManager.FinishScanRun(record, errStr)

		// Increment scan counters
		steamScanCnt++
		if steamScanCnt >= 55 {
			steamScanCnt = 0
		}
		csTradeScanCnt++
		if csTradeScanCnt >= 100 {
			csTradeScanCnt = 0
		}

		// Save back to DB
		dbState.SteamScanCount = steamScanCnt
		dbState.CsTradeScanCount = csTradeScanCnt
		r.dbClient.UpdateScannerRuntimeState(ctx, dbState)

		// Throttler (2000ms minimum cycle time)
		time.Sleep(2 * time.Second)
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
		notifs, err := sc.Scan(ctx)
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
		notifs, err := sc.Scan(ctx)
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
