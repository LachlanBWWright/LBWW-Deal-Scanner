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

	err = db.AutoMigrate(
		&SearchQuery{},
		&UserQuery{},
		&CashConverters{},
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
	conn, ok := c.conn.(driver.Pinger)
	if !ok {
		return driver.ErrSkip
	}
	return conn.Ping(ctx)
}

func (c *validatingConn) ResetSession(ctx context.Context) error {
	if conn, ok := c.conn.(driver.SessionResetter); ok {
		if err := conn.ResetSession(ctx); err != nil {
			return err
		}
	}

	conn, ok := c.conn.(driver.Pinger)
	if !ok {
		return driver.ErrBadConn
	}
	if err := conn.Ping(ctx); err != nil {
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
