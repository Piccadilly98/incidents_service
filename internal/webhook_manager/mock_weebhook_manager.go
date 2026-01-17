package webhook_manager

import (
	"sync"

	"github.com/Piccadilly98/incidents_service/internal/models/dto"
)

type MockWebhookManager struct {
	Storage []*dto.WebhookTask
	mu      sync.RWMutex
}

func NewMockWebhookManager() *MockWebhookManager {
	return &MockWebhookManager{}
}

func (mw *MockWebhookManager) AddToQueue(result dto.LocationCheckResponse, url string, method string) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.Storage = append(mw.Storage, &dto.WebhookTask{
		Dto:    result,
		Method: method,
		Url:    url,
	})
}

func (mw *MockWebhookManager) Stop() {}
