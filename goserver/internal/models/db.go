package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(connStr string, tursoAuthToken string) (*gorm.DB, error) {
	if connStr == "" {
		return nil, fmt.Errorf("TURSO_DATABASE_URL is required")
	}
	if tursoAuthToken == "" {
		return nil, fmt.Errorf("TURSO_AUTH_TOKEN is required")
	}
	if !strings.HasPrefix(connStr, "libsql://") && !strings.HasPrefix(connStr, "https://") {
		return nil, fmt.Errorf("TURSO_DATABASE_URL must start with libsql:// or https://")
	}

	connector, err := libsql.NewConnector(connStr, libsql.WithAuthToken(tursoAuthToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create libsql connector: %w", err)
	}

	sqlDB := sql.OpenDB(validatingConnector{connector: connector})
	// This is a remote libSQL database, so scanner and command traffic should
	// not be serialized through one physical connection.
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPing()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to connect to libsql database: %w", err)
	}

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to open libsql database: %w", err)
	}

	if err := migrateCashConvertersFilters(db); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to migrate Cash Converters filters: %w", err)
	}

	err = db.AutoMigrate(
		&SearchQuery{},
		&UserQuery{},
		&CashConverters{},
		&CashConvertersFilter{},
		&Ebay{},
		&Gumtree{},
		&Salvos{},
		&CsMarket{},
		&SteamMarket{},
		&CsTradeBot{},
		&Globals{},
		&Listing{},
		&ListingObservation{},
		&QueryListingState{},
		&ScannerRuntimeState{},
		&ActionRegistry{},
		&TtlItem{},
	)
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run GORM auto-migrations: %w", err)
	}

	return db, nil
}

func migrateCashConvertersFilters(db *gorm.DB) error {
	if !db.Migrator().HasTable("CashConverters") {
		return nil
	}
	if db.Migrator().HasTable("CashConvertersFilter") {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		type legacyCashConverters struct {
			Url                   string
			RequiredPhrases       string
			ExcludePhrases        string
			RequiredInDescription string
			ExcludeInDescription  string
			MinPrice              *float64
			MaxPrice              *float64
			ScanMode              string
			QueryId               string
		}

		var rows []legacyCashConverters
		if err := tx.Table("CashConverters").Find(&rows).Error; err != nil {
			return fmt.Errorf("read legacy rows: %w", err)
		}
		if err := tx.Migrator().RenameTable("CashConverters", "CashConvertersLegacy"); err != nil {
			return fmt.Errorf("rename legacy table: %w", err)
		}
		if err := tx.AutoMigrate(&CashConverters{}, &CashConvertersFilter{}); err != nil {
			return fmt.Errorf("create filter tables: %w", err)
		}

		for _, row := range rows {
			scanMode := row.ScanMode
			if scanMode == "" {
				scanMode = "searchUrl"
			}
			if err := tx.Create(&CashConverters{Url: row.Url, ScanMode: scanMode}).Error; err != nil {
				return fmt.Errorf("copy search %q: %w", row.Url, err)
			}
			required := combineLegacyPhrases(row.RequiredPhrases, row.RequiredInDescription)
			excluded := combineLegacyPhrases(row.ExcludePhrases, row.ExcludeInDescription)
			filter := CashConvertersFilter{
				ID:                row.QueryId,
				CashConvertersUrl: row.Url,
				QueryId:           row.QueryId,
				RequiredPhrases:   required,
				RequiredMatchMode: "all",
				ExcludePhrases:    excluded,
				ExcludeMatchMode:  "any",
				MinPrice:          row.MinPrice,
				MaxPrice:          row.MaxPrice,
				CreatedAt:         time.Now().UTC(),
			}
			if err := tx.Create(&filter).Error; err != nil {
				return fmt.Errorf("copy filter for %q: %w", row.Url, err)
			}
		}

		var copied int64
		if err := tx.Model(&CashConvertersFilter{}).Count(&copied).Error; err != nil {
			return fmt.Errorf("count copied filters: %w", err)
		}
		if copied != int64(len(rows)) {
			return fmt.Errorf("copied %d of %d Cash Converters filters", copied, len(rows))
		}
		return nil
	})
}

func combineLegacyPhrases(primary string, legacy string) string {
	primary = strings.TrimSpace(primary)
	legacy = strings.TrimSpace(legacy)
	if primary == "" {
		return legacy
	}
	if legacy == "" || legacy == primary {
		return primary
	}
	return primary + "," + legacy
}

// validatingConnector works around remote Hrana streams becoming closed while
// their database/sql connection remains in the idle pool. The upstream
// ResetSession implementation does not reject that connection, so validate it
// before database/sql hands it to another caller.
type validatingConnector struct {
	connector driver.Connector
}

func (c validatingConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &validatingConn{conn: conn}, nil
}

func (c validatingConnector) Driver() driver.Driver {
	return c.connector.Driver()
}

type validatingConn struct {
	conn driver.Conn
}

func (c *validatingConn) Prepare(query string) (driver.Stmt, error) {
	return c.conn.Prepare(query)
}

func (c *validatingConn) Close() error {
	return c.conn.Close()
}

func (c *validatingConn) Begin() (driver.Tx, error) {
	return c.conn.Begin()
}

func (c *validatingConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	conn, ok := c.conn.(driver.ConnPrepareContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return conn.PrepareContext(ctx, query)
}

func (c *validatingConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	conn, ok := c.conn.(driver.ConnBeginTx)
	if !ok {
		return nil, driver.ErrSkip
	}
	return conn.BeginTx(ctx, opts)
}

func (c *validatingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	conn, ok := c.conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return conn.ExecContext(ctx, query, args)
}

func (c *validatingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	conn, ok := c.conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return conn.QueryContext(ctx, query, args)
}

func (c *validatingConn) Ping(ctx context.Context) error {
	return c.validate(ctx)
}

func (c *validatingConn) ResetSession(ctx context.Context) error {
	if conn, ok := c.conn.(driver.SessionResetter); ok {
		if err := conn.ResetSession(ctx); err != nil {
			return err
		}
	}

	return c.validate(ctx)
}

func (c *validatingConn) validate(ctx context.Context) error {
	conn, ok := c.conn.(driver.ExecerContext)
	if !ok {
		return driver.ErrBadConn
	}
	if _, err := conn.ExecContext(ctx, "SELECT 1", nil); err != nil {
		return fmt.Errorf("validate pooled libsql connection: %w", err)
	}
	return nil
}

func (c *validatingConn) IsValid() bool {
	if conn, ok := c.conn.(driver.Validator); ok && !conn.IsValid() {
		return false
	}
	return true
}

func (c *validatingConn) CheckNamedValue(value *driver.NamedValue) error {
	conn, ok := c.conn.(driver.NamedValueChecker)
	if !ok {
		return driver.ErrSkip
	}
	return conn.CheckNamedValue(value)
}
