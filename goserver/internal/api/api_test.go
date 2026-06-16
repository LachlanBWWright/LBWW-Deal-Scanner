package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dealscanner/internal/config"
	"dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/runtime"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestApiStatusAndAuth(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}

	err = gdb.AutoMigrate(&models.Globals{}, &models.ScannerRuntimeState{})
	if err != nil {
		t.Fatalf("Failed to migrate test DB: %v", err)
	}
	dbClient := query.Use(gdb)

	cfg := &config.Config{
		ApiSecret:        "test-secret-key",
		EnableTestingApi: true,
	}

	stateManager := runtime.NewStateManager(3000)

	server := NewServer(cfg, dbClient, stateManager, nil, nil)

	t.Run("Unauthorized request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/status", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized (401), got %d", rec.Code)
		}
	})

	t.Run("Authorized request to status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/status", nil)
		req.Header.Set("X-API-SECRET", "test-secret-key")
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status OK (200), got %d", rec.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response JSON: %v", err)
		}

		apiStatus, ok := resp["api"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected api key in status response")
		}

		if apiStatus["port"] != float64(3000) {
			t.Errorf("Expected port 3000, got %v", apiStatus["port"])
		}
	})
}
