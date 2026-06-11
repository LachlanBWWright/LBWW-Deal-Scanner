package main

import (
	"log"

	"dealscanner/internal/config"
	"dealscanner/internal/models"
)

func main() {
	cfg := config.LoadConfig()

	db, err := models.Open(cfg.TursoDatabaseUrl, cfg.TursoAuthToken)
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to access SQL database handle: %v", err)
	}
	defer sqlDB.Close()

	log.Println("Turso database schema migration completed.")
}
