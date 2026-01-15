package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

type DbReposytory interface {
	Begin() (Tx, error)
	PingWithCtx(ctx context.Context) error
	PingWithTimeout(duration time.Duration) error
	Close() error
	RegistrationIncident(ctx context.Context, entit *entities.RegistrationIncidentEntitie, exec Executor) (string, error)
	GetInfoByIncidentID(ctx context.Context, id string, exec Executor) (*entities.ReadIncident, error)
	GetExistByIncidentID(ctx context.Context, id string, exec Executor) (bool, error)
	UpdateIncidentByID(ctx context.Context, id string, entit *entities.UpdateIncident, exec Executor) (*entities.ReadIncident, error)
	DeleteIncidentByID(ctx context.Context, id string, exec Executor) error
	GetCountRows(ctx context.Context, exec Executor) (int, error)
	GetPaginationIncidentsInfo(ctx context.Context, entit *entities.PaginationIncidents, exec Executor) ([]*entities.ReadIncident, error)
	RegistrationCheck(ctx context.Context, userID, latitude, longitude string, exec Executor) (string, error)
	GetDetectedIncidents(ctx context.Context, longitude, latitude string, exec Executor) ([]*entities.DistanceCheck, error)
	UpdateCheckByID(ctx context.Context, dangersIds []string, checkId string, isDanger bool, exec Executor) error
	GetCountUniqueUsers(ctx context.Context, exec Executor) (int, error)
	GetStaticsForIncidentsWithTimeWindow(ctx context.Context, exec Executor, timeWindow int) ([]*entities.IncidentStat, error)
	Name() string
}

type CacheReposytory interface {
	SetActiveIncident(ctx context.Context, data *entities.ReadIncident) error
	GetActiveIncident(ctx context.Context, id string) (*entities.ReadIncident, error)
	DeleteActiveIncident(ctx context.Context, id string) error
	PingWithCtx(ctx context.Context) error
	Name() string
}

type Executor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type CacheQueue interface {
	PopFromQueue(ctx context.Context) (*dto.WebhookTask, bool, error)
	AddToQueue(read *dto.WebhookTask, ctx context.Context) error
	PushTask(task *dto.WebhookTask, ctx context.Context) error
	PingWithCtx(ctx context.Context) error
	Name() string
}

type Tx interface {
	Executor
	Commit() error
	Rollback() error
}
