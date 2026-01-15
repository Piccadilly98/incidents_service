package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
	"github.com/google/uuid"
)

type MockDbRepository struct {
	Storage map[string]*entities.ReadIncident
	Mu      *sync.RWMutex
	Tx      *FakeTx
	InTx    bool
}

func NewMockDb() *MockDbRepository {
	return &MockDbRepository{
		Storage: make(map[string]*entities.ReadIncident),
		Mu:      &sync.RWMutex{}}
}

func (m *MockDbRepository) Begin() (Tx, error) {
	m.Tx = &FakeTx{}
	return m.Tx, nil
}

func (m *MockDbRepository) Close() error {
	return nil
}

func (m *MockDbRepository) PingWithTimeout(duration time.Duration) error {
	return nil
}

func (m *MockDbRepository) PingWithCtx(ctx context.Context) error {
	return nil
}

func (m *MockDbRepository) Name() string {
	return "mock_db_repository"
}

func (m *MockDbRepository) GetInfoByIncidentID(ctx context.Context, id string, exec Executor) (*entities.ReadIncident, error) {
	if exec != nil {
		m.InTx = true
	}
	m.Mu.RLock()
	stEntit, ok := m.Storage[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	m.Mu.RUnlock()
	res := &entities.ReadIncident{}
	res.Id = id
	res.Name = stEntit.Name
	res.Type = stEntit.Type
	res.Description = stEntit.Description
	res.Latitude = stEntit.Latitude
	res.Longitude = stEntit.Longitude
	res.Radius = stEntit.Radius
	res.IsActive = stEntit.IsActive
	res.Status = stEntit.Status
	res.Coordinates = fmt.Sprintf("POINT(%s %s)", stEntit.Longitude, stEntit.Latitude)
	res.UpdatedDate = &time.Time{}
	res.ResolvedDate = stEntit.ResolvedDate
	res.CreatedDate = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	return res, nil
}

func (m *MockDbRepository) GetExistByIncidentID(ctx context.Context, id string, exec Executor) (bool, error) {
	if exec != nil {
		m.InTx = true
	}
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	_, ok := m.Storage[id]
	return ok, nil
}

func (m *MockDbRepository) UpdateIncidentByID(ctx context.Context, id string, entit *entities.UpdateIncident, exec Executor) (*entities.ReadIncident, error) {
	if exec != nil {
		m.InTx = true
	}

	m.Mu.Lock()
	defer m.Mu.Unlock()

	res, ok := m.Storage[id]
	if !ok {
		return nil, sql.ErrNoRows
	}

	if entit.Type != nil {
		res.Type = *entit.Type
	}
	if entit.Status != nil {
		res.Status = *entit.Status
	}
	if entit.ResolvedTime != nil {
		res.ResolvedDate = entit.ResolvedTime
	}
	if entit.Radius != nil {
		res.Radius = *entit.Radius
	}
	if entit.Name != nil {
		res.Name = *entit.Name
	}
	res.IsActive = entit.IsActive
	if entit.Description != nil {
		res.Description = entit.Description
	}
	res.UpdatedDate = getTimePtr(time.Now().UTC())
	return res, nil
}

func (m *MockDbRepository) DeleteIncidentByID(ctx context.Context, id string, exec Executor) error {
	if exec != nil {
		m.InTx = true
	}

	m.Mu.Lock()
	defer m.Mu.Unlock()

	delete(m.Storage, id)
	return nil
}

func getTimePtr(t time.Time) *time.Time {
	return &t
}

func (m *MockDbRepository) GetCountRows(ctx context.Context, exec Executor) (int, error) {
	if exec != nil {
		m.InTx = true
	}
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	return len(m.Storage), nil
}

func (m *MockDbRepository) GetPaginationIncidentsInfo(ctx context.Context, entit *entities.PaginationIncidents, exec Executor) ([]*entities.ReadIncident, error) {
	return nil, nil
}

func (m *MockDbRepository) RegistrationCheck(ctx context.Context, userID, latitude, longitude string, exec Executor) (string, error) {
	return "", nil
}

func (m *MockDbRepository) GetDetectedIncidents(ctx context.Context, longitude, latitude string, exec Executor) ([]*entities.DistanceCheck, error) {
	return nil, nil
}

func (m *MockDbRepository) UpdateCheckByID(ctx context.Context, dangersIds []string, checkId string, isDanger bool, exec Executor) error {
	return nil
}

func (m *MockDbRepository) GetCountUniqueUsers(ctx context.Context, exec Executor) (int, error) {
	return 0, nil
}

func (m *MockDbRepository) GetStaticsForIncidentsWithTimeWindow(ctx context.Context, exec Executor, timeWindow int) ([]*entities.IncidentStat, error) {
	return nil, nil
}

func (m *MockDbRepository) RegistrationIncident(ctx context.Context, entit *entities.RegistrationIncidentEntitie, exec Executor) (string, error) {
	if exec != nil {
		m.InTx = true
	}
	uuid := uuid.NewString()
	m.Mu.Lock()
	defer m.Mu.Unlock()

	m.Storage[uuid] = &entities.ReadIncident{
		Id:           uuid,
		Name:         entit.Name,
		Type:         entit.Type,
		Description:  entit.Description,
		Latitude:     entit.Latitude,
		Longitude:    entit.Longitude,
		Radius:       entit.Radius,
		IsActive:     entit.IsActive,
		Status:       entit.Status,
		ResolvedDate: entit.ResolvedTime,
		CreatedDate:  time.Now().UTC(),
	}

	return uuid, nil
}
