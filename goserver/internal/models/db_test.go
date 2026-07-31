package models

import (
	"context"
	"database/sql/driver"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

type testConn struct {
	resetCalls int
	execCalls  int
	execErr    error
}

func (c *testConn) Prepare(string) (driver.Stmt, error) {
	return nil, driver.ErrSkip
}

func (c *testConn) Close() error {
	return nil
}

func (c *testConn) Begin() (driver.Tx, error) {
	return nil, driver.ErrSkip
}

func (c *testConn) ResetSession(context.Context) error {
	c.resetCalls++
	return nil
}

func (c *testConn) ExecContext(
	context.Context,
	string,
	[]driver.NamedValue,
) (driver.Result, error) {
	c.execCalls++
	return driver.RowsAffected(0), c.execErr
}

func TestValidatingConnResetsAndValidatesBeforeReuse(t *testing.T) {
	underlying := &testConn{}
	conn := &validatingConn{conn: underlying}

	if err := conn.ResetSession(context.Background()); err != nil {
		t.Fatalf("ResetSession() error = %v", err)
	}
	if underlying.resetCalls != 1 {
		t.Fatalf("ResetSession() reset calls = %d, want 1", underlying.resetCalls)
	}
	if underlying.execCalls != 1 {
		t.Fatalf("ResetSession() validation calls = %d, want 1", underlying.execCalls)
	}
}

func TestValidatingConnRejectsClosedStream(t *testing.T) {
	underlying := &testConn{
		execErr: errors.Join(errors.New("stream is closed"), driver.ErrBadConn),
	}
	conn := &validatingConn{conn: underlying}

	err := conn.ResetSession(context.Background())
	if !errors.Is(err, driver.ErrBadConn) {
		t.Fatalf("ResetSession() error = %v, want driver.ErrBadConn", err)
	}
}

func TestValidatingConnPingUsesValidationQuery(t *testing.T) {
	underlying := &testConn{}
	conn := &validatingConn{conn: underlying}

	if err := conn.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if underlying.execCalls != 1 {
		t.Fatalf("Ping() validation calls = %d, want 1", underlying.execCalls)
	}
}

func TestOpenLocalSQLitePersistsDataAndEnforcesForeignKeys(t *testing.T) {
	databaseURL := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "dealscanner.db"))

	first, err := Open(databaseURL, "")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	query := SearchQuery{ID: "query-1", CreatedAt: time.Now().UTC()}
	if err := first.Create(&query).Error; err != nil {
		t.Fatalf("create query: %v", err)
	}
	firstSQL, err := first.DB()
	if err != nil {
		t.Fatalf("first DB() error = %v", err)
	}
	if err := firstSQL.Close(); err != nil {
		t.Fatalf("close first database: %v", err)
	}

	second, err := Open(databaseURL, "")
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	secondSQL, err := second.DB()
	if err != nil {
		t.Fatalf("second DB() error = %v", err)
	}
	defer secondSQL.Close()

	var count int64
	if err := second.Model(&SearchQuery{}).Where("id = ?", query.ID).Count(&count).Error; err != nil {
		t.Fatalf("count persisted query: %v", err)
	}
	if count != 1 {
		t.Fatalf("persisted query count = %d, want 1", count)
	}

	invalidUserQuery := UserQuery{ID: "user-query-1", UserId: "user-1", QueryId: "missing"}
	if err := second.Create(&invalidUserQuery).Error; err == nil {
		t.Fatal("create user query without a parent query succeeded, want foreign key error")
	}
}

func TestOpenRejectsUnsupportedDatabaseURL(t *testing.T) {
	if _, err := Open("/var/lib/dealscanner/dealscanner.db", ""); err == nil {
		t.Fatal("Open() succeeded for an unsupported database URL")
	}
}

func TestExtractLegacyCashConvertersPhrase(t *testing.T) {
	tests := map[string]string{
		"https://www.cashconverters.com.au/search-results?query=nintendo%20switch": "nintendo switch",
		"https://www.cashconverters.com.au/shop/phones/iphone-15":                  "iphone 15",
		"  vintage camera  ": "vintage camera",
	}
	for input, want := range tests {
		if got := extractLegacyCashConvertersPhrase(input); got != want {
			t.Errorf("extractLegacyCashConvertersPhrase(%q) = %q, want %q", input, got, want)
		}
	}
}
