package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"dealscanner/internal/api"
	"dealscanner/internal/config"
	"dealscanner/internal/models"
	"dealscanner/internal/db/query"
	"dealscanner/internal/discord"
	"dealscanner/internal/notifications"
	"dealscanner/internal/runtime"
	"dealscanner/internal/scanners"
)

func main() {
	log.Println("Starting DealScanner Server in Go...")

	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Open DB
	gormDB, err := models.Open(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err == nil && sqlDB != nil {
		defer sqlDB.Close()
	}
	log.Println("Database connection established.")

	// Initialize GORM Gen query client
	queryClient := query.Use(gormDB)
	query.SetDefault(gormDB)

	// Seed default database globals from env variables if empty
	if err := queryClient.SeedGlobals(context.Background(), cfg); err != nil {
		log.Printf("Failed to seed default database globals: %v", err)
	}

	// 3. Initialize StateManager
	stateManager := runtime.NewStateManager(cfg.ApiPort)

	// 4. Initialize Discord Bot
	bot := discord.NewBot(cfg, queryClient, stateManager)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := bot.Start(ctx); err != nil {
		log.Printf("Failed to start Discord Bot: %v", err)
	}
	defer bot.Close()

	// 5. Initialize Notification Providers & Service
	discordProvider := notifications.NewDiscordProvider(cfg, bot)
	notifService := notifications.NewNotificationService([]notifications.NotificationProvider{
		discordProvider,
	})

	// 6. Initialize Scanner Runner and Site Scanners
	runner := scanners.NewRunner(cfg, queryClient, stateManager, notifService, bot)
	runner.RegisterScanner(scanners.NewEbayScanner(queryClient))
	runner.RegisterScanner(scanners.NewSteamMarketScanner(queryClient))
	runner.RegisterScanner(scanners.NewSalvosScanner(queryClient))
	runner.RegisterScanner(scanners.NewGumtreeScanner(queryClient))
	runner.RegisterScanner(scanners.NewCashConvertersScanner(queryClient))
	runner.RegisterScanner(scanners.NewCsTradeScanner(queryClient))
	runner.RegisterScanner(scanners.NewLootFarmScanner(queryClient))
	runner.RegisterScanner(scanners.NewTradeItScanner(queryClient))
	runner.RegisterScanner(scanners.NewCsMarketScanner(queryClient))

	// 7. Start Scanners background process
	runner.Start(ctx)

	// 8. Start HTTP API Server
	apiServer := api.NewServer(cfg, queryClient, stateManager, runner, notifService)
	addr := cfg.ApiHost + ":" + strconv.Itoa(cfg.ApiPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: apiServer,
	}

	go func() {
		log.Printf("HTTP Server listening on http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server failed: %v", err)
		}
	}()

	// 9. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	cancel() // cancel scanner background tasks

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP Server shutdown error: %v", err)
	}

	log.Println("Server stopped.")
}
