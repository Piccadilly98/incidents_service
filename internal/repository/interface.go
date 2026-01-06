package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

type DbReposytory interface {
	Begin() (*sql.Tx, error)
	PingWithCtx(ctx context.Context) error
	PingWithTimeout(duration time.Duration) error
	Close() error
	RegistrationIncident(ctx context.Context, entit *entities.RegistrationIncidentEntitie, exec Executor) (string, error)
	GetInfoByIncidentID(ctx context.Context, id string, exec Executor) (*entities.ReadIncident, error)
	GetExistByIncidentID(ctx context.Context, id string, exec Executor) (bool, error)
	UpdateIncidentByID(ctx context.Context, id string, entit *entities.UpdateIncident, exec Executor) (*entities.ReadIncident, error)
}

type CacheReposytory interface {
}

type Executor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
