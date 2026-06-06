package runtime

import (
	"sync"
	"time"
)

type BotStatus string

const (
	BotStatusStarting BotStatus = "starting"
	BotStatusReady    BotStatus = "ready"
	BotStatusDisabled BotStatus = "disabled"
	BotStatusError    BotStatus = "error"
)

type ScanStatus string

const (
	ScanStatusIdle    ScanStatus = "idle"
	ScanStatusRunning ScanStatus = "running"
	ScanStatusError   ScanStatus = "error"
)

type ScanRunRecord struct {
	Mode       string    `json:"mode"` // "loop" or "manual"
	StartedAt  string    `json:"startedAt"`
	FinishedAt *string   `json:"finishedAt"`
	DurationMs *int64    `json:"durationMs"`
	Error      *string   `json:"error"`
	startTime  time.Time // internal helper
}

type SearchResultRecord struct {
	Source    string   `json:"source"`
	Title     string   `json:"title"`
	Url       string   `json:"url"`
	Price     *float64 `json:"price"`
	ImageUrl  *string  `json:"imageUrl"`
	QueryType *string  `json:"queryType"`
	QueryId   *string  `json:"queryId"`
	FoundAt   string   `json:"foundAt"`
}

type ApiStatus struct {
	StartedAt string `json:"startedAt"`
	Port      int    `json:"port"`
}

type BotState struct {
	Status      BotStatus `json:"status"`
	Connected   bool      `json:"connected"`
	LastReadyAt *string   `json:"lastReadyAt"`
	LastError   *string   `json:"lastError"`
}

type ScannerState struct {
	Status        ScanStatus           `json:"status"`
	LoopRunning   bool                 `json:"loopRunning"`
	LastRun       *ScanRunRecord       `json:"lastRun"`
	RecentResults []SearchResultRecord `json:"recentResults"`
}

type RuntimeSnapshot struct {
	Api      ApiStatus    `json:"api"`
	Bot      BotState     `json:"bot"`
	Scanner  ScannerState `json:"scanner"`
	Commands []string     `json:"commands"`
}

type StateManager struct {
	mu    sync.RWMutex
	state RuntimeSnapshot
}

func NewStateManager(port int) *StateManager {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	return &StateManager{
		state: RuntimeSnapshot{
			Api: ApiStatus{
				StartedAt: nowStr,
				Port:      port,
			},
			Bot: BotState{
				Status:    BotStatusStarting,
				Connected: false,
			},
			Scanner: ScannerState{
				Status:        ScanStatusIdle,
				LoopRunning:   false,
				RecentResults: []SearchResultRecord{},
			},
			Commands: []string{},
		},
	}
}

func (s *StateManager) GetSnapshot() RuntimeSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Deep copy to prevent race condition on slices/pointers
	snap := RuntimeSnapshot{
		Api: s.state.Api,
		Bot: BotState{
			Status:    s.state.Bot.Status,
			Connected: s.state.Bot.Connected,
		},
		Scanner: ScannerState{
			Status:      s.state.Scanner.Status,
			LoopRunning: s.state.Scanner.LoopRunning,
		},
	}

	if s.state.Bot.LastReadyAt != nil {
		val := *s.state.Bot.LastReadyAt
		snap.Bot.LastReadyAt = &val
	}
	if s.state.Bot.LastError != nil {
		val := *s.state.Bot.LastError
		snap.Bot.LastError = &val
	}

	if s.state.Scanner.LastRun != nil {
		snap.Scanner.LastRun = &ScanRunRecord{
			Mode:      s.state.Scanner.LastRun.Mode,
			StartedAt: s.state.Scanner.LastRun.StartedAt,
		}
		if s.state.Scanner.LastRun.FinishedAt != nil {
			val := *s.state.Scanner.LastRun.FinishedAt
			snap.Scanner.LastRun.FinishedAt = &val
		}
		if s.state.Scanner.LastRun.DurationMs != nil {
			val := *s.state.Scanner.LastRun.DurationMs
			snap.Scanner.LastRun.DurationMs = &val
		}
		if s.state.Scanner.LastRun.Error != nil {
			val := *s.state.Scanner.LastRun.Error
			snap.Scanner.LastRun.Error = &val
		}
	}

	snap.Scanner.RecentResults = make([]SearchResultRecord, len(s.state.Scanner.RecentResults))
	copy(snap.Scanner.RecentResults, s.state.Scanner.RecentResults)

	snap.Commands = make([]string, len(s.state.Commands))
	copy(snap.Commands, s.state.Commands)

	return snap
}

func (s *StateManager) SetApiPort(port int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Api.Port = port
}

func (s *StateManager) SetBotStatus(status BotStatus, lastError *string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Bot.Status = status
	s.state.Bot.Connected = (status == BotStatusReady)
	s.state.Bot.LastError = lastError
	if status == BotStatusReady {
		nowStr := time.Now().UTC().Format(time.RFC3339)
		s.state.Bot.LastReadyAt = &nowStr
	}
}

func (s *StateManager) SetCommandNames(commands []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Commands = make([]string, len(commands))
	copy(s.state.Commands, commands)
}

func (s *StateManager) BeginScanLoop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Scanner.LoopRunning = true
}

func (s *StateManager) EndScanLoop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Scanner.LoopRunning = false
}

func (s *StateManager) BeginScanRun(mode string) *ScanRunRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	startedAt := now.Format(time.RFC3339)
	s.state.Scanner.Status = ScanStatusRunning

	record := &ScanRunRecord{
		Mode:      mode,
		StartedAt: startedAt,
		startTime: now,
	}
	s.state.Scanner.LastRun = record
	return record
}

func (s *StateManager) FinishScanRun(record *ScanRunRecord, errStr *string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record == nil {
		return
	}

	finishedAtTime := time.Now().UTC()
	finishedAt := finishedAtTime.Format(time.RFC3339)
	durationMs := finishedAtTime.Sub(record.startTime).Milliseconds()

	record.FinishedAt = &finishedAt
	record.DurationMs = &durationMs
	record.Error = errStr

	if errStr != nil {
		s.state.Scanner.Status = ScanStatusError
	} else {
		s.state.Scanner.Status = ScanStatusIdle
	}
}

func (s *StateManager) RecordSearchResults(results []SearchResultRecord) {
	if len(results) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	combined := append(results, s.state.Scanner.RecentResults...)
	if len(combined) > 100 {
		combined = combined[:100]
	}
	s.state.Scanner.RecentResults = combined
}
