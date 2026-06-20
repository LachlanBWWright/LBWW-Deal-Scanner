package models

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
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
