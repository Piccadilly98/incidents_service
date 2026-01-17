package service_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/config"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/models/entities"
	"github.com/Piccadilly98/incidents_service/internal/repository"
	"github.com/Piccadilly98/incidents_service/internal/service"
	"github.com/Piccadilly98/incidents_service/internal/webhook_manager"
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

func TestService_GetIncidentInfo(t *testing.T) {
	testCases := []struct {
		name               string
		existInCache       bool
		existInDb          bool
		checkId            string
		bodyInStorage      *entities.ReadIncident
		loadToCacheAfterDb bool
		expectBody         bool
		expectErr          bool
		containError       string
	}{
		{
			name:         "contain_in_cache_no_contain_in_db",
			existInCache: true,
			checkId:      "new_id",
			bodyInStorage: &entities.ReadIncident{
				Id:   "new_id",
				Name: "new_name",
				Type: "new_type",
			},
			expectBody: true,
		},
		{
			name:         "no_contain_in_cache_contain_in_db",
			existInCache: false,
			existInDb:    true,
			checkId:      "new_id",
			bodyInStorage: &entities.ReadIncident{
				Id:       "new_id",
				Name:     "new_name",
				Type:     "new_type",
				IsActive: true,
			},
			expectBody:         true,
			loadToCacheAfterDb: true,
		},
		{
			name:         "no_contain_in_cache_no_contain_in_db",
			existInCache: false,
			existInDb:    false,
			checkId:      "new_id",
			bodyInStorage: &entities.ReadIncident{
				Id:       "new_id",
				Name:     "new_name",
				Type:     "new_type",
				IsActive: true,
			},
			expectBody:         false,
			loadToCacheAfterDb: false,
			expectErr:          true,
			containError:       "no rows",
		},
		{
			name:         "no_contain_in_cache_get_not_active_not_load_in_cache",
			existInCache: false,
			existInDb:    true,
			checkId:      "new_id",
			bodyInStorage: &entities.ReadIncident{
				Id:       "new_id",
				Name:     "new_name",
				Type:     "new_type",
				IsActive: false,
			},
			expectBody:         true,
			loadToCacheAfterDb: false,
		},
		{
			name:         "data_from_db",
			existInCache: false,
			existInDb:    true,
			checkId:      "db-only-id",
			bodyInStorage: &entities.ReadIncident{
				Id:       "db-only-id",
				Name:     "Только в БД",
				Type:     "test",
				IsActive: true,
			},
			loadToCacheAfterDb: false,
			expectBody:         true,
			expectErr:          false,
		},
		{
			name:         "err_in_read_from_cache_read_from_db",
			existInCache: true,
			existInDb:    true,
			checkId:      "id-with-cache-error",
			bodyInStorage: &entities.ReadIncident{
				Id:       "id-with-cache-error",
				Name:     "Из БД после ошибки кэша",
				Type:     "test",
				IsActive: true,
			},
			loadToCacheAfterDb: true,
			expectBody:         true,
			expectErr:          false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDb := repository.NewMockDb()
			cacheMock := repository.NewCacheMock()
			if tc.existInCache {
				if tc.checkId == tc.bodyInStorage.Id {
					cacheMock.Storage[tc.bodyInStorage.Id] = tc.bodyInStorage
				}
			}
			if tc.existInDb {
				if tc.checkId == tc.bodyInStorage.Id {
					mockDb.Storage[tc.bodyInStorage.Id] = tc.bodyInStorage
				}
			}
			svc := service.NewService(mockDb, cacheMock, nil, nil)

			res, err := svc.GetIncidentInfoByID(context.Background(), tc.bodyInStorage.Id)
			if err != nil {
				if tc.expectErr {
					if !strings.Contains(err.Error(), tc.containError) {
						t.Errorf("ERROR: got: %s, expect: %s\n", err.Error(), tc.containError)
					}
				} else {
					t.Errorf("unexpected err: %s\n", err.Error())
				}

			}

			if tc.expectBody {
				if res != nil {
					if tc.checkId != res.ID {
						t.Errorf("ID: got: %s, expect: %s", res.ID, tc.checkId)
					}
					if tc.bodyInStorage.Name != res.Name {
						t.Errorf("NAME: got: %s, expect: %s\n", res.Name, tc.bodyInStorage.Name)
					}
					if tc.bodyInStorage.Type != res.Type {
						t.Errorf("TYPE: got: %s, expect: %s\n", res.Type, tc.bodyInStorage.Type)
					}
				} else {
					t.Errorf("RESULT:  got: not nil, expect: nil\n")
				}
				return
			} else {
				if res != nil {
					t.Errorf("unexpected body!\n")
				}
			}

			if tc.loadToCacheAfterDb {
				_, err := cacheMock.GetActiveIncident(context.Background(), tc.checkId)
				if err != nil {
					t.Errorf("CHECK EXISTS IN CACHE AFTER DB: got: true, expect: false\n")
				}
			} else {
				_, err := cacheMock.GetActiveIncident(context.Background(), tc.checkId)
				if err == nil {
					t.Errorf("CHECK EXISTS IN CACHE AFTER DB: got: false, expect: true\n")
				}
			}
		})
	}
}

func TestService_UpdateIncidentByID(t *testing.T) {
	testCases := []struct {
		name           string
		id             string
		bodyInStorage  *entities.ReadIncident
		containInDb    bool
		containInCache bool
		body           *dto.UpdateRequest
		expectInCache  bool
		expectErr      bool
		containErr     string
	}{
		{
			name:        "valid_update_type_no_cache",
			id:          "new_id",
			containInDb: true,
			bodyInStorage: &entities.ReadIncident{
				Id:       "new_id",
				Type:     "old",
				IsActive: true,
			},
			body: &dto.UpdateRequest{
				Type: getStrPtr("update_type"),
			},
			expectInCache: true,
		},

		{
			name:       "invalid_body_not_valid_no_data_for_update",
			id:         "new_id",
			body:       &dto.UpdateRequest{},
			expectErr:  true,
			containErr: "no data for update",
		},
		{
			name: "invalid_body_not_valid_name_empty",
			id:   "new_id",
			body: &dto.UpdateRequest{
				Name: getStrPtr(""),
			},
			expectErr:  true,
			containErr: "name cannot be empty",
		},
		{
			name:        "valid_update_radius_and_status",
			id:          "upd-id-1",
			containInDb: true,
			bodyInStorage: &entities.ReadIncident{
				Id:       "upd-id-1",
				Name:     "old_name",
				Type:     "old_type",
				Radius:   100,
				IsActive: true,
				Status:   "active",
			},
			body: &dto.UpdateRequest{
				Radius: getIntPtr(300),
				Status: getStrPtr("resolved"),
			},
			expectInCache: false,
			expectErr:     false,
		},

		{
			name:        "update_archived_incident_forbidden",
			id:          "archived-id",
			containInDb: true,
			bodyInStorage: &entities.ReadIncident{
				Id:       "archived-id",
				Name:     "old",
				IsActive: false,
				Status:   "archived",
			},
			body: &dto.UpdateRequest{
				Name: getStrPtr("new_name"),
			},
			expectErr:     true,
			containErr:    "unable to update archived incident",
			expectInCache: false,
		},

		{
			name:        "valid_partial_update_only_description",
			id:          "upd-id-2",
			containInDb: true,
			bodyInStorage: &entities.ReadIncident{
				Id:          "upd-id-2",
				Name:        "old_name",
				Description: getStrPtr("old_desc"),
				IsActive:    true,
			},
			body: &dto.UpdateRequest{
				Description: getStrPtr("new_description"),
			},
			expectInCache: true,
			expectErr:     false,
		},

		{
			name:        "no_changes_in_update → ошибка",
			id:          "upd-id-3",
			containInDb: true,
			bodyInStorage: &entities.ReadIncident{
				Id:       "upd-id-3",
				Name:     "test",
				Type:     "test",
				IsActive: true,
			},
			body: &dto.UpdateRequest{
				Name: getStrPtr("test"), // то же самое имя
			},
			expectErr:     true,
			containErr:    "no data for update",
			expectInCache: false,
		},

		{
			name:        "update_inactive_incident_to_active",
			id:          "upd-id-4",
			containInDb: true,
			bodyInStorage: &entities.ReadIncident{
				Id:       "upd-id-4",
				Name:     "old",
				IsActive: false,
				Status:   "resolved",
			},
			body: &dto.UpdateRequest{
				Status: getStrPtr("active"),
			},
			expectInCache: true,
			expectErr:     false,
		},

		{
			name:        "radius_out_of_max_limit → ошибка",
			id:          "upd-id-5",
			containInDb: true,
			bodyInStorage: &entities.ReadIncident{
				Id:       "upd-id-5",
				Radius:   100,
				IsActive: true,
			},
			body: &dto.UpdateRequest{
				Radius: getIntPtr(6000),
			},
			expectErr:     true,
			containErr:    "radius cannot be >",
			expectInCache: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDb := repository.NewMockDb()
			mockCache := repository.NewCacheMock()
			cfg := &config.Config{
				DefaultRadius: 500,
				MaxRadius:     5000,
			}
			svc := service.NewService(mockDb, mockCache, cfg, nil)
			if tc.containInCache {
				if tc.bodyInStorage != nil {
					mockCache.Storage[tc.bodyInStorage.Id] = tc.bodyInStorage
				} else {
					t.Fatalf("invalid test: body in storage cannot be nil and contain in cache: true")
				}
			}
			if tc.containInDb {
				if tc.bodyInStorage != nil {
					mockDb.Storage[tc.bodyInStorage.Id] = tc.bodyInStorage
				} else {
					t.Fatalf("invalid test: body in storage cannot be nil and contain in db: true")
				}
			}

			res, err := svc.UpdateIncidentByID(context.Background(), tc.id, tc.body)
			if err != nil {
				if tc.expectErr {
					if !strings.Contains(err.Error(), tc.containErr) {
						t.Errorf("ERROR not contains %s in %s\n", tc.containErr, err.Error())
					}
				} else {
					t.Errorf("unexpected err: got: %s, expect: nil\n", err.Error())
				}

				if res != nil {
					t.Errorf("unexpected result %v\n", *res)
				}

				if mockDb.InTx {
					if mockDb.Tx != nil {
						if !mockDb.Tx.RolledBack {
							t.Errorf("tx cannot rollback!\n")
						}
					} else {
						t.Errorf("invalid mock method, tx cannot be start and inTx = true")
					}
				}
				return
			}

			if tc.expectInCache {
				_, err := mockCache.GetActiveIncident(context.Background(), tc.id)
				if err != nil {
					t.Errorf("error exists in cache: got: false, expect: true\n")
				}
			}

			if mockDb.Tx != nil {
				if !mockDb.Tx.Committed {
					t.Errorf("tx cannot commit\n")
				} else if mockDb.Tx.RolledBack {
					t.Errorf("unexpected rollback in tx\n")
				}
			}

			if tc.body != nil {
				if tc.body.Name != nil {
					if res.Name != *tc.body.Name {
						t.Errorf("fail update name: got: %s, expect: %s\n", res.Name, *tc.body.Name)
					}
				}
				if tc.body.Type != nil {
					if res.Type != *tc.body.Type {
						t.Errorf("fail update type: got: %s, expect: %s\n", res.Type, *tc.body.Type)
					}
				}
			}
		})
	}
}

func TestService_DeactivateIncidentByID(t *testing.T) {
	testCases := []struct {
		name           string
		id             string
		bodyInStorage  *entities.ReadIncident
		containInDb    bool
		containInCache bool
		expectInCache  bool
		expectErr      bool
		containErr     string
	}{
		{
			name: "invalid_deactivate_arhived_in_cache",
			id:   "new_id",
			bodyInStorage: &entities.ReadIncident{
				Id:     "new_id",
				Status: service.StatusArchived,
			},
			containInDb: true,
			expectErr:   true,
			containErr:  "incident already archived",
		},
		{
			name:           "data_only_cache_error_sql",
			id:             "cache-only-id",
			containInDb:    false,
			containInCache: true,
			bodyInStorage: &entities.ReadIncident{
				Id:       "cache-only-id",
				Name:     "Только в кэше",
				IsActive: true,
				Status:   service.StatusActive,
			},
			expectInCache: true,
			expectErr:     true,
			containErr:    "no rows",
		},

		{
			name: "success_deactivate_status_active",
			id:   "new_id",
			bodyInStorage: &entities.ReadIncident{
				Id:     "new_id",
				Status: service.StatusActive,
			},
			containInDb: true,
		},
		{
			name: "success_deactivate_status_resolved",
			id:   "new_id",
			bodyInStorage: &entities.ReadIncident{
				Id:     "new_id",
				Status: service.StatusResolved,
			},
			containInDb: true,
		},
		{
			name: "success_deactivate_status_active_contain_cache",
			id:   "new_id",
			bodyInStorage: &entities.ReadIncident{
				Id:     "new_id",
				Status: service.StatusActive,
			},
			containInDb:    true,
			containInCache: true,
		},

		{
			name:           "data_in_cache_and_db_success",
			id:             "active-cache-id",
			containInDb:    true,
			containInCache: true,
			bodyInStorage: &entities.ReadIncident{
				Id:       "active-cache-id",
				Name:     "Активный с кэшем",
				IsActive: true,
				Status:   service.StatusActive,
			},
			expectInCache: false,
			expectErr:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkTx := false
			mockDb := repository.NewMockDb()
			mockCache := repository.NewCacheMock()
			cfg := &config.Config{
				DefaultRadius: 500,
				MaxRadius:     5000,
			}
			svc := service.NewService(mockDb, mockCache, cfg, nil)
			if tc.containInCache {
				if tc.bodyInStorage != nil {
					mockCache.Storage[tc.bodyInStorage.Id] = tc.bodyInStorage
				} else {
					t.Fatalf("invalid test: body in storage cannot be nil and contain in cache: true")
				}
			}
			if tc.containInDb {
				if tc.bodyInStorage != nil {
					mockDb.Storage[tc.bodyInStorage.Id] = tc.bodyInStorage
				} else {
					t.Fatalf("invalid test: body in storage cannot be nil and contain in db: true")
				}
			}
			res, err := svc.DeactivateIncidentByID(context.Background(), tc.id)
			if err != nil {
				if tc.expectErr {
					if !strings.Contains(err.Error(), tc.containErr) {
						t.Errorf("ERROR not contains %s in %s\n", tc.containErr, err.Error())
					}
				} else {
					t.Errorf("unexpected err: got: %s, expect: nil\n", err.Error())
				}

				if res != nil {
					t.Errorf("unexpected result %v\n", *res)
				}

				if mockDb.InTx {
					if mockDb.Tx != nil {
						if !mockDb.Tx.RolledBack {
							t.Errorf("tx cannot rollback!\n")
						}
					} else {
						t.Errorf("invalid mock method, tx cannot be start and inTx = true")
					}
				}
				checkTx = true
			}
			if !checkTx {
				if mockDb.Tx != nil {
					if !mockDb.Tx.Committed {
						t.Errorf("tx cannot commit\n")
					} else if mockDb.Tx.RolledBack {
						t.Errorf("unexpected rollback in tx\n")
					}
				}
			}

			if tc.expectInCache {
				_, err := mockCache.GetActiveIncident(context.Background(), tc.id)
				if err != nil {
					t.Errorf("incident not contains in cache")
				}
			} else {
				_, err := mockCache.GetActiveIncident(context.Background(), tc.id)
				if err == nil {
					t.Errorf("incident not delete from cache")
				}
			}
			if res != nil {
				if res.Status != service.StatusArchived {
					t.Errorf("STATUS: got: %s, expect: %s\n", res.Status, service.StatusArchived)
				}
				if res.IsActive {
					t.Errorf("is_active: got: true, expect: false\n")
				}
				if res.ResolvedDate == nil {
					t.Errorf("ResolvedDate cannot be nil")
				}
				if res.UpdatedDate == nil {
					t.Errorf("UpdatedDate cannot be nil")
				}
			}
		})
	}
}

func TestService_DeleteIncidentByID(t *testing.T) {
	loadId := "new_id"
	body := &entities.ReadIncident{
		Id:          loadId,
		Type:        "type",
		Description: getStrPtr("new_type"),
		Latitude:    "38.8",
		Longitude:   "38.8",
		Radius:      100,
	}
	testCases := []struct {
		name              string
		deleteID          string
		containInDb       bool
		containInCache    bool
		expectBodyInCache bool
		expectBodyInDb    bool
	}{
		{
			name:           "delete_from_cache_and_db",
			deleteID:       loadId,
			containInDb:    true,
			containInCache: true,
		},
		{
			name:           "delete_only_cache",
			deleteID:       loadId,
			containInCache: true,
		},
		{
			name:        "delete_only_db",
			deleteID:    loadId,
			containInDb: true,
		},
		{
			name:              "delete_random_id",
			deleteID:          "random",
			containInDb:       true,
			containInCache:    true,
			expectBodyInCache: true,
			expectBodyInDb:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDb := repository.NewMockDb()
			mockCache := repository.NewCacheMock()
			svc := service.NewService(mockDb, mockCache, nil, nil)

			if tc.containInCache {

				mockCache.Storage[loadId] = body
			}
			if tc.containInDb {
				mockDb.Storage[loadId] = body
			}

			err := svc.DeleteIncidentByID(context.Background(), tc.deleteID)
			if err != nil {
				t.Errorf("unexpected err: %s\n", err.Error())
			}

			if mockDb.Tx != nil {
				t.Errorf("unexpected tx in db\n")
			}
			if !tc.expectBodyInCache {
				if _, err := mockDb.GetInfoByIncidentID(context.Background(), loadId, nil); err == nil {
					t.Errorf("unexpected row in db with id: %s\n", loadId)
				}
			} else {
				if _, err := mockDb.GetInfoByIncidentID(context.Background(), loadId, nil); err != nil {
					t.Errorf("invalid delete, cannot contains in cache\n")
				}
			}
			if !tc.expectBodyInDb {
				if _, err := mockCache.GetActiveIncident(context.Background(), loadId); err == nil {
					t.Errorf("unexpected row in cache with id: %s\n", loadId)
				}
			} else {
				if _, err := mockCache.GetActiveIncident(context.Background(), loadId); err != nil {
					t.Errorf("invalid delete, cannot contains in db\n")
				}
			}
		})
	}
}

func TestService_LocationCheck(t *testing.T) {
	inc1 := &entities.ReadIncident{
		Id:        "inc_1",
		Latitude:  "55.755826",
		Longitude: "37.617300",
		Status:    service.StatusActive,
		IsActive:  true,
		Radius:    2000,
	}
	inc2 := &entities.ReadIncident{
		Id:        "inc_2",
		Latitude:  "55.7735",
		Longitude: "37.6200",
		Status:    service.StatusActive,
		IsActive:  true,
		Radius:    1000,
	}
	inc3 := &entities.ReadIncident{
		Id:          "inc_3",
		Latitude:    "37.1406",
		Longitude:   "115.4840",
		Status:      service.StatusActive,
		IsActive:    true,
		Description: getStrPtr("zone-51!"),
		Radius:      1000,
	}
	inc4 := &entities.ReadIncident{
		Id:          "inc_4",
		Latitude:    "37",
		Longitude:   "115",
		Status:      service.StatusArchived,
		IsActive:    false,
		Description: getStrPtr("inactive"),
		Radius:      1000,
	}
	testCases := []struct {
		name                  string
		checkBody             *dto.LocationCheckRequest
		expectCheck           bool
		expectedIds           []string
		expectedIsDanger      bool
		expectedErrContain    string
		ContainsIdsInWeebhook bool
	}{
		{
			name: "check_coord_inc_1",
			checkBody: &dto.LocationCheckRequest{
				UserID:    "new_user",
				Latitude:  "55.755826",
				Longitude: "37.617300",
			},
			expectCheck:      true,
			expectedIds:      []string{"inc_1"},
			expectedIsDanger: true,
		},
		{
			name: "check_coord_inc_1_and_inc_2",
			checkBody: &dto.LocationCheckRequest{
				UserID:    "test_user",
				Latitude:  "55.7650",
				Longitude: "37.6184",
			},
			expectCheck:      true,
			expectedIds:      []string{"inc_1", "inc_2"},
			expectedIsDanger: true,
		},
		{
			name: "check_inactive_incident",
			checkBody: &dto.LocationCheckRequest{
				UserID:    "test_user",
				Latitude:  "37",
				Longitude: "115",
			},
			expectCheck:      true,
			expectedIsDanger: false,
		},

		{
			name: "invalid_user_id_empty",
			checkBody: &dto.LocationCheckRequest{
				UserID:    "",
				Latitude:  "37",
				Longitude: "115",
			},
			expectCheck:        false,
			expectedIsDanger:   false,
			expectedErrContain: "user_id cannot be empty",
		},
		{
			name: "invalid_latitude_empty",
			checkBody: &dto.LocationCheckRequest{
				UserID:    "user_1",
				Latitude:  "",
				Longitude: "115",
			},
			expectCheck:        false,
			expectedIsDanger:   false,
			expectedErrContain: "latitude incorrect parse",
		},

		{
			name: "вне всех зон",
			checkBody: &dto.LocationCheckRequest{
				UserID:    "user-out",
				Latitude:  "55.9000",
				Longitude: "37.8000",
			},
			expectedIsDanger: false,
			expectCheck:      true,
			expectedIds:      []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDb := repository.NewMockDb()
			mockCache := repository.NewCacheMock()
			mockWebhook := webhook_manager.NewMockWebhookManager()
			svc := service.NewService(mockDb, mockCache, nil, mockWebhook)

			mockDb.Storage[inc1.Id] = inc1
			mockDb.Storage[inc2.Id] = inc2
			mockDb.Storage[inc3.Id] = inc3
			mockDb.Storage[inc4.Id] = inc4

			res, err := svc.LocationCheck(context.Background(), tc.checkBody)

			if err != nil {
				if tc.expectedErrContain != "" {
					if !strings.Contains(err.Error(), tc.expectedErrContain) {
						t.Errorf("ERROR: got: %s, expect: %s\n", err.Error(), tc.expectedErrContain)
					}
				} else {
					t.Errorf("unexpected error: %s\n", err.Error())
				}
			}
			if tc.expectCheck {
				if _, ok := mockDb.Checks[res.ID]; !ok {
					t.Errorf("check_id: %s, not contains in checks\n", res.ID)
				}
				if mockDb.Tx != nil {
					if !mockDb.Tx.Committed {
						t.Errorf("tx cannot commit\n")
					}
				} else {
					t.Errorf("tx cannot start")
				}
			} else {
				if mockDb.Tx != nil {
					t.Errorf("tx cannot be started\n")
				}
				if res != nil {
					t.Errorf("unexpected result\n")
				}
			}
			if res != nil {
				if res.IsDanger != tc.expectedIsDanger {
					t.Errorf("is_danger: got: %v, expect: %v\n", res.IsDanger, tc.expectedIsDanger)
				}
				if res.Latitude != tc.checkBody.Latitude {
					t.Errorf("result latitude != request latitude\n")
				}
				if res.Longitude != tc.checkBody.Longitude {
					t.Errorf("result lingitude != request longitude\n")
				}
				if res.UserID != tc.checkBody.UserID {
					t.Errorf("result user_id != request user_id\n")
				}
				if tc.ContainsIdsInWeebhook {
					if len(res.DetectedIncidentsID) != len(tc.expectedIds) {
						t.Errorf("COUNT DETECTED: got: %d, expect: %d\n", len(res.DetectedIncidentsID), len(tc.expectedIds))
					}
					if len(mockWebhook.Storage) != 1 {
						t.Errorf("webhook not add new task\n")
					}
					if len(mockWebhook.Storage) == 1 {
						if len(mockWebhook.Storage[0].Dto.DetectedIncidentsID) != len(tc.expectedIds) {
							t.Errorf("COUNT DETECTED IN WEBHOOK: got: %d, expect: %d\n", len(mockWebhook.Storage[0].Dto.DetectedIncidentsID), len(tc.expectedIds))
						}
					}
				}
				for _, incident := range res.DetectedIncidentsID {
					if !slices.Contains(tc.expectedIds, incident.ID) {
						t.Errorf("not contains id in result: %s\n", incident.ID)
					}
				}

			} else {
				if len(mockWebhook.Storage) != 0 {
					t.Errorf("unexpected webhook sends\n")
				}
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

func getIntPtr(i int) *int {
	return &i
}
