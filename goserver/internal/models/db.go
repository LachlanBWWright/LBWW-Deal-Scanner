package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(connStr string, tursoAuthToken string) (*gorm.DB, error) {
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if strings.HasPrefix(connStr, "file:") {
		return openLocalSQLite(connStr)
	}
	if strings.HasPrefix(connStr, "libsql://") || strings.HasPrefix(connStr, "https://") {
		return openTurso(connStr, tursoAuthToken)
	}

	return nil, fmt.Errorf("DATABASE_URL must start with file:, libsql://, or https://")
}

func openLocalSQLite(connStr string) (*gorm.DB, error) {
	sqlDB, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("open local SQLite database: %w", err)
	}

	// SQLite still serializes writers, but WAL mode allows API reads to proceed
	// while the scanner is doing bounded work. A small pool prevents one
	// long-running scanner query from monopolizing the only connection without
	// creating a stampede of competing SQLite writers.
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)

	if err := configureLocalSQLite(sqlDB); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return openGormAndMigrate(sqlDB, "local SQLite")
}

func configureLocalSQLite(sqlDB *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, pragma := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA synchronous = NORMAL",
	} {
		if _, err := sqlDB.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("configure local SQLite with %q: %w", pragma, err)
		}
	}

	return nil
}

func openTurso(connStr string, tursoAuthToken string) (*gorm.DB, error) {
	if tursoAuthToken == "" {
		return nil, fmt.Errorf("TURSO_AUTH_TOKEN is required when DATABASE_URL selects Turso")
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

	return openGormAndMigrate(sqlDB, "libsql")
}

func openGormAndMigrate(sqlDB *sql.DB, backend string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to open %s database: %w", backend, err)
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
		&CashConvertersScanState{},
		&CashConvertersListingMeta{},
		&CashConvertersDetailJob{},
		&CashConvertersSearchDoc{},
		&CashConvertersDeletedListing{},
	)
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run GORM auto-migrations: %w", err)
	}

	if err := migrateCashConvertersSiteWideQueries(db); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to migrate legacy Cash Converters queries: %w", err)
	}

	if err := migrateCashConvertersSearch(db); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to migrate Cash Converters search index: %w", err)
	}

	return db, nil
}

func migrateCashConvertersSearch(db *gorm.DB) error {
	var documentCount int64
	if err := db.Model(&CashConvertersSearchDoc{}).Count(&documentCount).Error; err != nil {
		return fmt.Errorf("count Cash Converters search documents: %w", err)
	}

	// Backfill is a one-time migration. Running the INSERT OR IGNORE against the
	// complete listings table on every process start makes a large local SQLite
	// database appear unavailable while it scans millions of rows and grows the
	// WAL. Once documents exist, normal writes keep them current.
	if documentCount == 0 {
		backfillSQL := `
		INSERT OR IGNORE INTO CashConvertersSearchDoc (
			listingId, title, description, canonicalUrl, imageUrl, totalPrice,
			availability, sourceStatus, lastSeenAt, lastDetailAt, updatedAt
		)
		SELECT
			l.id,
			l.title,
			COALESCE(l.description, ''),
			l.canonicalUrl,
			l.imageUrl,
			(
				SELECT o.totalPrice
				FROM ListingObservation o
				WHERE o.listingId = l.id
				ORDER BY o.observedAt DESC
				LIMIT 1
			),
			COALESCE(l.availability, 'available'),
			COALESCE(m.sourceStatus, 'available'),
			l.lastSeenAt,
			l.lastDetailAt,
			CURRENT_TIMESTAMP
		FROM Listing l
		LEFT JOIN CashConvertersListingMeta m ON m.listingId = l.id
		WHERE l.source = 'cashConverters'
		`
		if err := db.Exec(backfillSQL).Error; err != nil {
			return fmt.Errorf("backfill Cash Converters search documents: %w", err)
		}
	}

	if err := migrateCashConvertersFTS(db); err != nil {
		// FTS5 is optional; LIKE search still uses the backfilled search documents.
		fmt.Printf("Warning: CashConvertersSearchFts migration failed (using LIKE fallback): %v\n", err)
	}

	for _, indexSQL := range []string{
		"CREATE INDEX IF NOT EXISTS idx_cc_search_doc_updated ON CashConvertersSearchDoc(updatedAt)",
		"CREATE INDEX IF NOT EXISTS idx_cc_search_doc_seen ON CashConvertersSearchDoc(lastSeenAt)",
		"CREATE INDEX IF NOT EXISTS idx_cc_search_available_seen ON CashConvertersSearchDoc(availability, sourceStatus, lastSeenAt DESC)",
		"CREATE INDEX IF NOT EXISTS idx_cc_search_available_price ON CashConvertersSearchDoc(availability, sourceStatus, totalPrice)",
		"CREATE INDEX IF NOT EXISTS idx_cc_meta_missing_due ON CashConvertersListingMeta(sourceStatus, missingVerifiedAt, missingVerificationCount, updatedAt)",
		"CREATE INDEX IF NOT EXISTS idx_cc_observation_latest ON ListingObservation(listingId, observedAt DESC)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_cc_detail_stage ON CashConvertersDetailJob(listingId, stage)",
		"CREATE INDEX IF NOT EXISTS idx_cc_detail_queue_priority ON CashConvertersDetailJob(completedAt, priority DESC, dueAt)",
	} {
		if err := db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("create Cash Converters index: %w", err)
		}
	}
	return nil
}

func migrateCashConvertersFTS(db *gorm.DB) error {
	var existingSQL string
	row := db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'CashConvertersSearchFts'").Row()
	err := row.Scan(&existingSQL)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("inspect Cash Converters FTS table: %w", err)
	}

	isExternalContent := strings.Contains(existingSQL, "content='CashConvertersSearchDoc'")
	if existingSQL != "" && !isExternalContent {
		if err := db.Exec("DROP TABLE CashConvertersSearchFts").Error; err != nil {
			return fmt.Errorf("replace legacy Cash Converters FTS table: %w", err)
		}
	}

	statements := []string{
		`CREATE VIRTUAL TABLE IF NOT EXISTS CashConvertersSearchFts USING fts5(
			listingId UNINDEXED,
			title,
			description,
			canonicalUrl UNINDEXED,
			content='CashConvertersSearchDoc',
			content_rowid='rowid',
			tokenize='unicode61 remove_diacritics 2'
		)`,
		`CREATE TRIGGER IF NOT EXISTS cc_search_doc_ai AFTER INSERT ON CashConvertersSearchDoc BEGIN
			INSERT INTO CashConvertersSearchFts(rowid, listingId, title, description, canonicalUrl)
			VALUES (new.rowid, new.listingId, new.title, new.description, new.canonicalUrl);
		END`,
		`CREATE TRIGGER IF NOT EXISTS cc_search_doc_ad AFTER DELETE ON CashConvertersSearchDoc BEGIN
			INSERT INTO CashConvertersSearchFts(CashConvertersSearchFts, rowid, listingId, title, description, canonicalUrl)
			VALUES ('delete', old.rowid, old.listingId, old.title, old.description, old.canonicalUrl);
		END`,
		`DROP TRIGGER IF EXISTS cc_search_doc_au`,
		`CREATE TRIGGER cc_search_doc_au AFTER UPDATE OF title, description, listingId, canonicalUrl ON CashConvertersSearchDoc
		WHEN old.title IS NOT new.title OR old.description IS NOT new.description
			OR old.listingId IS NOT new.listingId OR old.canonicalUrl IS NOT new.canonicalUrl
		BEGIN
			INSERT INTO CashConvertersSearchFts(CashConvertersSearchFts, rowid, listingId, title, description, canonicalUrl)
			VALUES ('delete', old.rowid, old.listingId, old.title, old.description, old.canonicalUrl);
			INSERT INTO CashConvertersSearchFts(rowid, listingId, title, description, canonicalUrl)
			VALUES (new.rowid, new.listingId, new.title, new.description, new.canonicalUrl);
		END`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("create Cash Converters FTS index: %w", err)
		}
	}
	if existingSQL == "" || !isExternalContent {
		if err := db.Exec("INSERT INTO CashConvertersSearchFts(CashConvertersSearchFts) VALUES ('rebuild')").Error; err != nil {
			return fmt.Errorf("rebuild Cash Converters FTS index: %w", err)
		}
	}
	return nil
}

func migrateCashConvertersSiteWideQueries(db *gorm.DB) error {
	type legacyFilter struct {
		ID                string
		CashConvertersURL string `gorm:"column:cashConvertersUrl"`
		RequiredPhrases   string
		ScanMode          string
	}

	var filters []legacyFilter
	if err := db.Table("CashConvertersFilter AS f").
		Select("f.id, f.cashConvertersUrl, f.requiredPhrases, c.scanMode").
		Joins("JOIN CashConverters AS c ON c.url = f.cashConvertersUrl").
		Where("c.scanMode = ? OR c.scanMode = ?", "", "searchUrl").
		Find(&filters).Error; err != nil {
		return fmt.Errorf("load legacy Cash Converters queries: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, filter := range filters {
			if strings.TrimSpace(filter.RequiredPhrases) == "" {
				phrase := extractLegacyCashConvertersPhrase(filter.CashConvertersURL)
				if phrase != "" {
					if err := tx.Model(&CashConvertersFilter{}).
						Where("id = ?", filter.ID).
						Update("requiredPhrases", phrase).Error; err != nil {
						return fmt.Errorf("migrate Cash Converters phrases for %s: %w", filter.ID, err)
					}
				}
			}
			if err := tx.Model(&CashConverters{}).
				Where("url = ?", filter.CashConvertersURL).
				Update("scanMode", "siteWide").Error; err != nil {
				return fmt.Errorf("migrate Cash Converters scan mode for %s: %w", filter.ID, err)
			}
		}
		return nil
	})
}

func extractLegacyCashConvertersPhrase(rawURL string) string {
	trimmed := strings.TrimSpace(rawURL)
	parsed, err := url.Parse(trimmed)
	if err == nil && parsed.Host != "" {
		for _, key := range []string{"query", "q", "search"} {
			if phrase := strings.TrimSpace(parsed.Query().Get(key)); phrase != "" && phrase != "0" {
				return phrase
			}
		}
		pathParts := strings.FieldsFunc(parsed.Path, func(r rune) bool {
			return r == '/'
		})
		if len(pathParts) > 0 {
			last := strings.TrimSpace(strings.NewReplacer("-", " ", "_", " ").Replace(pathParts[len(pathParts)-1]))
			if last != "" && last != "shop" && last != "search" {
				return last
			}
		}
		return ""
	}
	if !strings.Contains(trimmed, "://") {
		return trimmed
	}
	return ""
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
			if scanMode == "" || scanMode == "searchUrl" {
				scanMode = "siteWide"
			}
			if err := tx.Create(&CashConverters{Url: row.Url, ScanMode: scanMode}).Error; err != nil {
				return fmt.Errorf("copy search %q: %w", row.Url, err)
			}
			required := combineLegacyPhrases(row.RequiredPhrases, row.RequiredInDescription)
			if required == "" {
				required = extractLegacyCashConvertersPhrase(row.Url)
			}
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
