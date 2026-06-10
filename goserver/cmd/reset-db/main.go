package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"dealscanner/internal/config"
	"dealscanner/internal/models"
	"github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	cfg := config.LoadConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if cfg.TursoDatabaseUrl == "" {
		log.Fatal("TURSO_DATABASE_URL is required")
	}
	if cfg.TursoAuthToken == "" {
		log.Fatal("TURSO_AUTH_TOKEN is required")
	}

	if err := dropUserTables(ctx, cfg.TursoDatabaseUrl, cfg.TursoAuthToken); err != nil {
		log.Fatalf("Failed to reset database: %v", err)
	}

	db, err := models.Open(cfg.TursoDatabaseUrl, cfg.TursoAuthToken)
	if err != nil {
		log.Fatalf("Failed to recreate Go schema: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to access SQL database handle: %v", err)
	}
	defer sqlDB.Close()

	log.Println("Turso database reset and Go schema migration completed.")
}

func dropUserTables(ctx context.Context, databaseUrl string, authToken string) error {
	connector, err := libsql.NewConnector(databaseUrl, libsql.WithAuthToken(authToken))
	if err != nil {
		return fmt.Errorf("create libsql connector: %w", err)
	}

	db := sql.OpenDB(connector)
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("read table name: %w", err)
		}
		tableNames = append(tableNames, tableName)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tables: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}

	for _, tableName := range tableNames {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", quoteIdent(tableName))
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("drop table %s: %w", tableName, err)
		}
		log.Printf("Dropped table %s", tableName)
	}

	return nil
}

func quoteIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
