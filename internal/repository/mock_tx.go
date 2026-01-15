package repository

import (
	"context"
	"database/sql"
)

type FakeTx struct {
	Committed  bool
	RolledBack bool
	Closed     bool
}

func (f *FakeTx) Commit() error {
	f.Committed = true
	return nil
}

func (f *FakeTx) Rollback() error {
	if !f.Committed {
		f.RolledBack = true
	}
	return nil
}

func (f *FakeTx) Close() error {
	f.Closed = true
	return nil
}

func (f *FakeTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (f *FakeTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (f *FakeTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
