package models

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
)

type testConn struct {
	resetCalls int
	pingCalls  int
	pingErr    error
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

func (c *testConn) Ping(context.Context) error {
	c.pingCalls++
	return c.pingErr
}

func TestValidatingConnResetsAndPingsBeforeReuse(t *testing.T) {
	underlying := &testConn{}
	conn := &validatingConn{conn: underlying}

	if err := conn.ResetSession(context.Background()); err != nil {
		t.Fatalf("ResetSession() error = %v", err)
	}
	if underlying.resetCalls != 1 {
		t.Fatalf("ResetSession() reset calls = %d, want 1", underlying.resetCalls)
	}
	if underlying.pingCalls != 1 {
		t.Fatalf("ResetSession() ping calls = %d, want 1", underlying.pingCalls)
	}
}

func TestValidatingConnRejectsClosedStream(t *testing.T) {
	underlying := &testConn{
		pingErr: errors.Join(errors.New("stream is closed"), driver.ErrBadConn),
	}
	conn := &validatingConn{conn: underlying}

	err := conn.ResetSession(context.Background())
	if !errors.Is(err, driver.ErrBadConn) {
		t.Fatalf("ResetSession() error = %v, want driver.ErrBadConn", err)
	}
}
