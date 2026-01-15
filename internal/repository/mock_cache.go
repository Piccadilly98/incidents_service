package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

/*
SetActiveIncident(ctx context.Context, data *entities.ReadIncident) error
	GetActiveIncident(ctx context.Context, id string) (*entities.ReadIncident, error)
	DeleteActiveIncident(ctx context.Context, id string) error
	PingWithCtx(ctx context.Context) error
*/

type CacheMock struct {
	Storage map[string]*entities.ReadIncident
	mu      sync.RWMutex
}

func NewCacheMock() *CacheMock {
	return &CacheMock{
		Storage: make(map[string]*entities.ReadIncident),
	}
}

func (cm *CacheMock) SetActiveIncident(ctx context.Context, data *entities.ReadIncident) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.Storage[data.Id] = data
	return nil
}

func (cm *CacheMock) GetActiveIncident(ctx context.Context, id string) (*entities.ReadIncident, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	res, ok := cm.Storage[id]
	if !ok {
		return nil, fmt.Errorf("no contains with id")
	}

	return res, nil
}

func (cm *CacheMock) DeleteActiveIncident(ctx context.Context, id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.Storage, id)
	return nil
}

func (cm *CacheMock) PingWithCtx(ctx context.Context) error {
	return nil
}

func (cm *CacheMock) Name() string {
	return "cache-mock"
}
