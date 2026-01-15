package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/config"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/repository"
	"github.com/Piccadilly98/incidents_service/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestService_RegistrationIncident_No_Cache(t *testing.T) {
	type testCase struct {
		name           string
		req            *dto.RegistrationIncidentRequest
		wantErr        bool
		wantErrContain string
		wantID         bool
		checkStorage   bool
		checkTx        bool
	}

	tests := []testCase{
		{
			name: "успешное создание без ошибок",
			req: &dto.RegistrationIncidentRequest{
				Name:           "Пожар в лесу",
				Type:           "fire",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				RadiusInMeters: ptrInt(500),
			},
			wantErr:      false,
			wantID:       true,
			checkStorage: true,
			checkTx:      true,
		},
		{
			name: "ошибка валидации - пустое имя",
			req: &dto.RegistrationIncidentRequest{
				Name:      "",
				Type:      "fire",
				Latitude:  "55.7558",
				Longitude: "37.6173",
			},
			wantErr:        true,
			wantErrContain: "name cannot be empty",
			checkStorage:   false,
			checkTx:        false,
		},
		{
			name: "ошибка валидации - неверный радиус",
			req: &dto.RegistrationIncidentRequest{
				Name:           "Тест",
				Type:           "test",
				Latitude:       "55.0",
				Longitude:      "37.0",
				RadiusInMeters: ptrInt(-10),
			},
			wantErr:        true,
			wantErrContain: "radius cannot be <= 0",
			checkStorage:   false,
			checkTx:        false,
		},
		{
			name: "ошибка валидации - длинное имя",
			req: &dto.RegistrationIncidentRequest{
				Name:           strings.Repeat("a", 101),
				Type:           "test",
				Latitude:       "55.0",
				Longitude:      "37.0",
				RadiusInMeters: ptrInt(100),
			},
			wantErr:        true,
			wantErrContain: "very long name",
			checkStorage:   false,
			checkTx:        false,
		},
		{
			name: "ошибка валидации - длинный тип",
			req: &dto.RegistrationIncidentRequest{
				Name:           "name",
				Type:           strings.Repeat("a", 101),
				Latitude:       "55.0",
				Longitude:      "37.0",
				RadiusInMeters: ptrInt(100),
			},
			wantErr:        true,
			wantErrContain: "very long type",
			checkStorage:   false,
			checkTx:        false,
		},
		{
			name: "ошибка валидации - неверный статус",
			req: &dto.RegistrationIncidentRequest{
				Name:           "name",
				Type:           "type",
				Latitude:       "55.0",
				Longitude:      "37.0",
				Status:         getStrPtr("unexpected"),
				RadiusInMeters: ptrInt(100),
			},
			wantErr:        true,
			wantErrContain: "invalid status",
			checkStorage:   false,
			checkTx:        false,
		},
		{
			name: "ошибка валидации - длинный статус",
			req: &dto.RegistrationIncidentRequest{
				Name:           "name",
				Type:           "type",
				Latitude:       "55.0",
				Longitude:      "37.0",
				Status:         getStrPtr(strings.Repeat("a", 21)),
				RadiusInMeters: ptrInt(100),
			},
			wantErr:        true,
			wantErrContain: "very long status",
			checkStorage:   false,
			checkTx:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockDb()
			cfg := &config.Config{
				DefaultRadius: 500,
				MaxRadius:     5000,
			}
			svc := service.NewService(mockRepo, nil, cfg, nil)
			res, err := svc.RegistrationIncident(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				if tt.wantID {
					assert.NotEmpty(t, res.ID)
				}
			}

			if tt.checkStorage {
				assert.Len(t, mockRepo.Storage, 1)
				stored, exists := mockRepo.Storage[res.ID]
				assert.True(t, exists)
				assert.Equal(t, tt.req.Name, stored.Name)
				assert.Equal(t, tt.req.Type, stored.Type)
				assert.Equal(t, tt.req.Latitude, stored.Latitude)
				assert.True(t, stored.CreatedDate.After(time.Now().Add(-time.Minute)))
			} else {
				assert.Empty(t, mockRepo.Storage)
			}

			if tt.checkTx {
				assert.NotNil(t, mockRepo.Tx)
				assert.True(t, mockRepo.Tx.Committed, "транзакция должна быть закоммичена")
				assert.False(t, mockRepo.Tx.RolledBack, "rollback не должен вызываться")
			}
		})
	}
}

func TestService_RegistrationIncident_WithCache(t *testing.T) {
	type testCase struct {
		name           string
		req            *dto.RegistrationIncidentRequest
		wantCacheEntry bool
		wantErr        bool
		wantErrContain string
	}

	tests := []testCase{
		{
			name: "успешное создание + активный инцидент → сохраняется в кэш",
			req: &dto.RegistrationIncidentRequest{
				Name:           "Пожар в лесу",
				Type:           "fire",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				RadiusInMeters: ptrInt(500),
				Status:         getStrPtr("active"),
			},
			wantCacheEntry: true,
			wantErr:        false,
		},
		{
			name: "успешное создание + кэш выключен → кэш не трогается",
			req: &dto.RegistrationIncidentRequest{
				Name:      "Тест без кэша",
				Type:      "test",
				Latitude:  "55.0",
				Longitude: "37.0",
			},
			wantCacheEntry: false,
			wantErr:        false,
		},
		{
			name: "ошибка валидации → кэш не трогается",
			req: &dto.RegistrationIncidentRequest{
				Name:      "",
				Type:      "fire",
				Latitude:  "55.7558",
				Longitude: "37.6173",
			},
			wantCacheEntry: false,
			wantErr:        true,
			wantErrContain: "name cannot be empty",
		},
		{
			name: "создание неактивного инцидента → не сохраняется в кэш",
			req: &dto.RegistrationIncidentRequest{
				Name:      "Архивный инцидент",
				Type:      "archived",
				Latitude:  "55.0",
				Longitude: "37.0",
				Status:    getStrPtr("archived"),
			},
			wantCacheEntry: false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockDb()
			mockCache := repository.NewCacheMock()

			cfg := &config.Config{
				DefaultRadius: 500,
				MaxRadius:     5000,
			}

			svc := service.NewService(mockRepo, mockCache, cfg, nil)

			// Act
			res, err := svc.RegistrationIncident(context.Background(), tt.req)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.NotEmpty(t, res.ID)
			}

			if tt.wantCacheEntry {
				cached, cacheErr := mockCache.GetActiveIncident(context.Background(), res.ID)
				assert.NoError(t, cacheErr)
				assert.NotNil(t, cached)
				assert.Equal(t, tt.req.Name, cached.Name)
				assert.True(t, cached.IsActive)
			}
		})
	}
}

func ptrInt(i int) *int {
	return &i
}

func getStrPtr(s string) *string {
	return &s
}
