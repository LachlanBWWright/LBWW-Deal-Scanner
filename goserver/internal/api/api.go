package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dealscanner/internal/config"
	"dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/notifications"
	"dealscanner/internal/runtime"
	"dealscanner/internal/scanners"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	cfg          *config.Config
	dbClient     *query.Query
	stateManager *runtime.StateManager
	runner       *scanners.Runner
	notifService *notifications.NotificationService
	router       *chi.Mux

	// In-memory testing logs matching testingState.ts
	testMu              sync.RWMutex
	notificationHistory []map[string]interface{}
	scanHistory         []map[string]interface{}
}

func NewServer(cfg *config.Config, dbClient *query.Query, stateManager *runtime.StateManager, runner *scanners.Runner, notifService *notifications.NotificationService) *Server {
	s := &Server{
		cfg:          cfg,
		dbClient:     dbClient,
		stateManager: stateManager,
		runner:       runner,
		notifService: notifService,
		router:       chi.NewRouter(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// CORS middleware
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-SECRET, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Auth Middleware
	s.router.Use(s.verifyApiSecret)

	// API routes
	s.router.Route("/api", func(r chi.Router) {
		r.Get("/status", s.handleStatus)
		r.Get("/commands", s.handleCommands)
		r.Post("/scans/run", s.handleRunScan)
		r.Get("/search-results", s.handleSearchResults)
		r.Get("/queries", s.handleListQueries)
		r.Post("/queries", s.handleCreateQuery)
		r.Put("/queries", s.handleUpdateQuery)
		r.Delete("/queries", s.handleDeleteQuery)

		// Testing API
		r.Route("/testing", func(tr chi.Router) {
			tr.Use(s.requireTestingEnabled)
			tr.Get("/capabilities", s.handleTestingCapabilities)
			tr.Post("/notifications", s.handleSendTestNotification)
			tr.Get("/notifications", s.handleGetTestNotifications)
			tr.Post("/scans/run", s.handleTestingRunScan)
			tr.Get("/runs", s.handleGetTestingRuns)
		})
	})

	// Serve Frontend Static Files
	distDir := s.resolveDistDir()
	if distDir != "" {
		s.setupFileServer(distDir)
	}
}

func (s *Server) verifyApiSecret(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api") {
			next.ServeHTTP(w, r)
			return
		}

		if s.cfg.ApiSecret == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Server misconfiguration"}`))
			return
		}

		provided := r.Header.Get("X-API-SECRET")
		if provided == "" {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				provided = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		provided = strings.TrimSpace(provided)

		if provided == "" || provided != s.cfg.ApiSecret {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireTestingEnabled(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.cfg.EnableTestingApi {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "Testing API is not enabled"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Handler functions

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	snap := s.stateManager.GetSnapshot()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap)
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	snap := s.stateManager.GetSnapshot()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"commands": snap.Commands})
}

func (s *Server) handleRunScan(w http.ResponseWriter, r *http.Request) {
	s.runner.RunManualScanOnce(r.Context())
	snap := s.stateManager.GetSnapshot()
	var startedAt, finishedAt string
	var durationMs int64
	var mode string = "manual"
	var errMsg *string

	if snap.Scanner.LastRun != nil {
		startedAt = snap.Scanner.LastRun.StartedAt
		if snap.Scanner.LastRun.FinishedAt != nil {
			finishedAt = *snap.Scanner.LastRun.FinishedAt
		}
		if snap.Scanner.LastRun.DurationMs != nil {
			durationMs = *snap.Scanner.LastRun.DurationMs
		}
		mode = snap.Scanner.LastRun.Mode
		errMsg = snap.Scanner.LastRun.Error
	} else {
		startedAt = time.Now().UTC().Format(time.RFC3339)
		finishedAt = startedAt
	}

	response := map[string]interface{}{
		"accepted":   true,
		"startedAt":  startedAt,
		"finishedAt": finishedAt,
		"durationMs": durationMs,
		"mode":       mode,
		"error":      errMsg,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSearchResults(w http.ResponseWriter, r *http.Request) {
	qType := r.URL.Query().Get("type")
	qId := r.URL.Query().Get("queryId")

	snap := s.stateManager.GetSnapshot()
	var results []runtime.SearchResultRecord

	for _, item := range snap.Scanner.RecentResults {
		typeMatches := qType == "" || (item.QueryType != nil && *item.QueryType == qType)
		queryMatches := qId == "" || (item.QueryId != nil && *item.QueryId == qId)
		if typeMatches && queryMatches {
			results = append(results, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"results": results})
}

func (s *Server) handleListQueries(w http.ResponseWriter, r *http.Request) {
	qType := r.URL.Query().Get("type")
	queries, err := s.dbClient.ListSavedQueries(r.Context(), qType)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"queries": queries})
}

func (s *Server) handleCreateQuery(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	var dmStruct struct {
		DmOnly bool `json:"dmOnly"`
	}
	_ = json.Unmarshal(body.Payload, &dmStruct)
	dmOnly := dmStruct.DmOnly

	var savedQuery query.QueryItem
	var err error

	switch body.Type {
	case "cashConverters":
		var cc query.CashConvertersQueryInput
		if err = json.Unmarshal(body.Payload, &cc); err == nil {
			savedQuery, err = s.dbClient.CreateCashConvertersQuery(r.Context(), dmOnly, cc)
		}
	case "ebay":
		var eb models.Ebay
		if err = json.Unmarshal(body.Payload, &eb); err == nil {
			savedQuery, err = s.dbClient.CreateEbayQuery(r.Context(), dmOnly, &eb)
		}
	case "gumtree":
		var gt models.Gumtree
		if err = json.Unmarshal(body.Payload, &gt); err == nil {
			savedQuery, err = s.dbClient.CreateGumtreeQuery(r.Context(), dmOnly, &gt)
		}
	case "salvos":
		var sa models.Salvos
		if err = json.Unmarshal(body.Payload, &sa); err == nil {
			savedQuery, err = s.dbClient.CreateSalvosQuery(r.Context(), dmOnly, &sa)
		}
	case "csMarket":
		var cm models.CsMarket
		if err = json.Unmarshal(body.Payload, &cm); err == nil {
			savedQuery, err = s.dbClient.CreateCsMarketQuery(r.Context(), dmOnly, &cm)
		}
	case "steamMarket":
		var sm models.SteamMarket
		if err = json.Unmarshal(body.Payload, &sm); err == nil {
			savedQuery, err = s.dbClient.CreateSteamMarketQuery(r.Context(), dmOnly, &sm)
		}
	case "csTradeBot":
		var ct models.CsTradeBot
		if err = json.Unmarshal(body.Payload, &ct); err == nil {
			savedQuery, err = s.dbClient.CreateCsTradeBotQuery(r.Context(), dmOnly, &ct)
		}
	default:
		err = fmt.Errorf("unknown query type: %s", body.Type)
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"query":   savedQuery,
	})
}

func (s *Server) handleUpdateQuery(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type    string          `json:"type"`
		Id      string          `json:"id"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	var dmStruct struct {
		DmOnly bool `json:"dmOnly"`
	}
	_ = json.Unmarshal(body.Payload, &dmStruct)
	dmOnly := dmStruct.DmOnly

	var savedQuery query.QueryItem
	var err error

	switch body.Type {
	case "cashConverters":
		var cc query.CashConvertersQueryInput
		if err = json.Unmarshal(body.Payload, &cc); err == nil {
			savedQuery, err = s.dbClient.UpdateCashConvertersQuery(r.Context(), body.Id, dmOnly, cc)
		}
	case "ebay":
		var eb models.Ebay
		if err = json.Unmarshal(body.Payload, &eb); err == nil {
			savedQuery, err = s.dbClient.UpdateEbayQuery(r.Context(), body.Id, dmOnly, &eb)
		}
	case "gumtree":
		var gt models.Gumtree
		if err = json.Unmarshal(body.Payload, &gt); err == nil {
			savedQuery, err = s.dbClient.UpdateGumtreeQuery(r.Context(), body.Id, dmOnly, &gt)
		}
	case "salvos":
		var sa models.Salvos
		if err = json.Unmarshal(body.Payload, &sa); err == nil {
			savedQuery, err = s.dbClient.UpdateSalvosQuery(r.Context(), body.Id, dmOnly, &sa)
		}
	case "csMarket":
		var cm models.CsMarket
		if err = json.Unmarshal(body.Payload, &cm); err == nil {
			savedQuery, err = s.dbClient.UpdateCsMarketQuery(r.Context(), body.Id, dmOnly, &cm)
		}
	case "steamMarket":
		var sm models.SteamMarket
		if err = json.Unmarshal(body.Payload, &sm); err == nil {
			savedQuery, err = s.dbClient.UpdateSteamMarketQuery(r.Context(), body.Id, dmOnly, &sm)
		}
	case "csTradeBot":
		var ct models.CsTradeBot
		if err = json.Unmarshal(body.Payload, &ct); err == nil {
			savedQuery, err = s.dbClient.UpdateCsTradeBotQuery(r.Context(), body.Id, dmOnly, &ct)
		}
	default:
		err = fmt.Errorf("unknown query type: %s", body.Type)
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"query":   savedQuery,
	})
}

func (s *Server) handleDeleteQuery(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type string `json:"type"`
		Id   string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	err := s.dbClient.DeleteSavedQuery(r.Context(), body.Type, body.Id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// Testing API Handlers

func (s *Server) handleTestingCapabilities(w http.ResponseWriter, r *http.Request) {
	snap := s.stateManager.GetSnapshot()
	queryTypes := []string{"cashConverters", "ebay", "gumtree", "salvos", "csMarket", "steamMarket", "csTradeBot"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"testingEnabled":        s.cfg.EnableTestingApi,
		"notificationProviders": []string{"discord"},
		"discordConnected":      snap.Bot.Connected,
		"availableQueryTypes":   queryTypes,
		"availableScannerTypes": queryTypes,
	})
}

func (s *Server) handleSendTestNotification(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Kind     string   `json:"kind"`
		Source   string   `json:"source"`
		Title    string   `json:"title"`
		Url      string   `json:"url"`
		Price    *float64 `json:"price"`
		ImageUrl string   `json:"imageUrl"`
		Message  string   `json:"message"`
		Query    *struct {
			Type   string `json:"type"`
			Id     string `json:"id"`
			DmOnly bool   `json:"dmOnly"`
		} `json:"query"`
		Tags                []string `json:"tags"`
		DeliveryMode        string   `json:"deliveryMode"`
		TargetDiscordUserId string   `json:"targetDiscordUserId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	startedAt := time.Now().UTC()
	id := fmt.Sprintf("notif-%d-%s", time.Now().UnixNano(), "rand")

	// build AppNotification
	notif := notifications.AppNotification{
		Kind:    body.Kind,
		Source:  body.Source,
		Title:   body.Title,
		Url:     body.Url,
		Price:   body.Price,
		Message: body.Message,
		Tags:    body.Tags,
	}
	if body.ImageUrl != "" {
		notif.ImageUrl = &body.ImageUrl
	}
	if body.Query != nil {
		notif.Query = &notifications.NotificationQuery{
			Type:   body.Query.Type,
			Id:     body.Query.Id,
			DmOnly: body.Query.DmOnly,
		}
	}

	s.notifService.Publish(r.Context(), notif)

	finishedAt := time.Now().UTC()
	durationMs := finishedAt.Sub(startedAt).Milliseconds()

	outcomes := []map[string]interface{}{
		{
			"provider": "all",
			"status":   "sent",
		},
	}

	result := map[string]interface{}{
		"id":         id,
		"startedAt":  startedAt.Format(time.RFC3339),
		"finishedAt": finishedAt.Format(time.RFC3339),
		"durationMs": durationMs,
		"request":    body,
		"outcomes":   outcomes,
		"error":      nil,
	}

	s.testMu.Lock()
	s.notificationHistory = append([]map[string]interface{}{result}, s.notificationHistory...)
	if len(s.notificationHistory) > 100 {
		s.notificationHistory = s.notificationHistory[:100]
	}
	s.testMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"startedAt":  startedAt.Format(time.RFC3339),
		"finishedAt": finishedAt.Format(time.RFC3339),
		"durationMs": durationMs,
		"outcomes":   outcomes,
		"error":      nil,
	})
}

func (s *Server) handleGetTestNotifications(w http.ResponseWriter, r *http.Request) {
	s.testMu.RLock()
	history := make([]map[string]interface{}, len(s.notificationHistory))
	copy(history, s.notificationHistory)
	s.testMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"results": history})
}

func (s *Server) handleTestingRunScan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type    string                 `json:"type"`
		Payload map[string]interface{} `json:"payload"`
		Notify  bool                   `json:"notify"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	startedAt := time.Now().UTC()
	id := fmt.Sprintf("scan-%d-%s", time.Now().UnixNano(), "rand")

	scanRes := scanners.RunTemporaryScan(r.Context(), body.Type, body.Payload)
	if scanRes.TimedOut {
		s.notifService.Publish(context.WithoutCancel(r.Context()), notifications.AppNotification{
			Kind:    "error",
			Source:  body.Type,
			Message: fmt.Sprintf("Warning: scanner %q exceeded the %s time limit. The scan was cancelled.", body.Type, scanners.DefaultScanTimeout),
			Tags:    []string{"warning", "scan-timeout"},
		})
	}

	if body.Notify && len(scanRes.Notifications) > 0 {
		for _, notif := range scanRes.Notifications {
			s.notifService.Publish(r.Context(), notif)
		}
		scanRes.NotificationsPublished = true
	}

	finishedAt := time.Now().UTC()
	durationMs := finishedAt.Sub(startedAt).Milliseconds()

	result := map[string]interface{}{
		"id":                     id,
		"startedAt":              startedAt.Format(time.RFC3339),
		"finishedAt":             finishedAt.Format(time.RFC3339),
		"durationMs":             durationMs,
		"request":                body,
		"items":                  scanRes.Items,
		"notifications":          scanRes.Notifications,
		"errors":                 scanRes.Errors,
		"timedOut":               scanRes.TimedOut,
		"notificationsPublished": scanRes.NotificationsPublished,
	}

	s.testMu.Lock()
	s.scanHistory = append([]map[string]interface{}{result}, s.scanHistory...)
	if len(s.scanHistory) > 100 {
		s.scanHistory = s.scanHistory[:100]
	}
	s.testMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                     id,
		"startedAt":              startedAt.Format(time.RFC3339),
		"finishedAt":             finishedAt.Format(time.RFC3339),
		"durationMs":             durationMs,
		"items":                  scanRes.Items,
		"notifications":          scanRes.Notifications,
		"errors":                 scanRes.Errors,
		"timedOut":               scanRes.TimedOut,
		"notificationsPublished": scanRes.NotificationsPublished,
	})
}

func (s *Server) handleGetTestingRuns(w http.ResponseWriter, r *http.Request) {
	s.testMu.RLock()
	history := make([]map[string]interface{}, len(s.scanHistory))
	copy(history, s.scanHistory)
	s.testMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"results": history})
}

// File server helpers

func (s *Server) resolveDistDir() string {
	candidates := []string{
		"../frontend/dist",
		"./frontend/dist",
		"../../frontend/dist",
	}

	for _, cand := range candidates {
		info, err := os.Stat(cand)
		if err == nil && info.IsDir() {
			abs, err := filepath.Abs(cand)
			if err == nil {
				return abs
			}
		}
	}
	return ""
}

func (s *Server) setupFileServer(distDir string) {
	fs := http.FileServer(http.Dir(distDir))

	s.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// If requesting api/docs/openapi paths that are not found, respond with 404
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/docs") || strings.HasPrefix(r.URL.Path, "/openapi") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Not Found"}`))
			return
		}

		// If the file exists, serve it
		path := filepath.Join(distDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA router support
		http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
	})
}
