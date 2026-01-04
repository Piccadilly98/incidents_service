package service

import (
	"strings"
	"testing"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/config"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

func TestService_ProcessingStatus(t *testing.T) {
	testCases := []struct {
		name           string
		requestStatus  *string
		expectedStatus string
		expectedError  string
	}{
		{
			name:           "nil status returns active",
			requestStatus:  nil,
			expectedStatus: StatusActive,
			expectedError:  "",
		},
		{
			name:           "active status returns active",
			requestStatus:  getPtrStr(StatusActive),
			expectedStatus: StatusActive,
			expectedError:  "",
		},
		{
			name:           "resolved status returns resolved",
			requestStatus:  getPtrStr(StatusResolved),
			expectedStatus: StatusResolved,
			expectedError:  "",
		},
		{
			name:           "invalid status returns error",
			requestStatus:  getPtrStr("invalid_status"),
			expectedStatus: "",
			expectedError:  "invalid status",
		},
		{
			name:           "empty string status returns error",
			requestStatus:  getPtrStr(""),
			expectedStatus: "",
			expectedError:  "invalid status",
		},
		{
			name:           "pending status returns error (if not allowed)",
			requestStatus:  getPtrStr("pending"),
			expectedStatus: "",
			expectedError:  "invalid status",
		},
	}

	s := &Service{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &dto.RegistrationIncidentRequest{
				Status: tc.requestStatus,
			}

			status, err := s.processingStatus(req)
			if err != nil {
				if tc.expectedError != err.Error() {
					t.Errorf("ERROR: got: %s, expect: %s\n", err.Error(), tc.expectedError)
				}
				if status != "" {
					t.Errorf("unexpected status: %s\n", status)
				}
				return
			} else {
				if tc.expectedError != "" {
					t.Errorf("ERROR: got nil, expect: %s\n", tc.expectedError)
					return
				}
			}
			if status != tc.expectedStatus {
				t.Errorf("STATUS: got: %s, expect: %s\n", status, tc.expectedStatus)
			}
		})
	}
}

func TestService_processingRadius(t *testing.T) {
	cfg := config.Config{
		DefaultRadius: 5000,
		MaxRadius:     10000,
	}
	testCases := []struct {
		name           string
		requestRadius  *int
		expectedRadius int
		expectedError  string
	}{
		{
			name:           "valid_1",
			requestRadius:  getIntPtr(100),
			expectedRadius: 100,
		},
		{
			name:           "edge_low",
			requestRadius:  getIntPtr(1),
			expectedRadius: 1,
		},
		{
			name:           "edge_high",
			requestRadius:  getIntPtr(9999),
			expectedRadius: 9999,
		},

		{
			name:           "invalid_radius_zero",
			requestRadius:  getIntPtr(0),
			expectedRadius: 0,
			expectedError:  "radius cannot be <= 0",
		},
		{
			name:           "invalid_radius_negative",
			requestRadius:  getIntPtr(-1),
			expectedRadius: 0,
			expectedError:  "radius cannot be <= 0",
		},

		{
			name:           "invalid_radius > max",
			requestRadius:  getIntPtr(10001),
			expectedRadius: 0,
			expectedError:  "radius cannot be >",
		},
		{
			name:           "invalid_radius over > max",
			requestRadius:  getIntPtr(10000000),
			expectedRadius: 0,
			expectedError:  "radius cannot be",
		},

		{
			name:           "valid_radius_nil",
			expectedRadius: cfg.DefaultRadius,
		},
	}
	s := Service{config: &cfg}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &dto.RegistrationIncidentRequest{
				RadiusInMeters: tc.requestRadius,
			}
			res, err := s.processingRadius(req)
			if err != nil {
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("ERROR: got: %s, expect: %s\n", err.Error(), tc.expectedError)
				}
				if res != 0 {
					t.Errorf("unexpected radius: %d\n", res)
				}
			} else {
				if tc.expectedError != "" {
					t.Errorf("ERROR: got nil, expect: %s\n", tc.expectedError)
					return
				}
			}

			if res != tc.expectedRadius {
				t.Errorf("RADIUS: got: %d, expect: %d\n", res, tc.expectedRadius)
			}
		})
	}
}

func TestService_processingIsActive(t *testing.T) {
	testCases := []struct {
		name           string
		status         string
		expectedResult bool
		expectedError  string
	}{
		{
			name:           "status_active",
			status:         StatusActive,
			expectedResult: true,
		},
		{
			name:           "status_resolved",
			status:         StatusResolved,
			expectedResult: false,
		},
		{
			name:           "invalid_status_empty",
			status:         "",
			expectedResult: false,
			expectedError:  "unexpected status",
		},
		{
			name:           "invalid_status_random",
			status:         "random_status",
			expectedResult: false,
			expectedError:  "unexpected status",
		},
		{
			name:           "invalid_status_pending",
			status:         "pending",
			expectedResult: false,
			expectedError:  "unexpected status",
		},
	}

	s := &Service{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := s.processingIsActive(tc.status)
			if err != nil {
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("ERROR: got: %v, expect to contain: %s", err, tc.expectedError)
				}
			} else {
				if tc.expectedError != "" {
					t.Errorf("ERROR: got nil, expect: %s", tc.expectedError)
					return
				}
			}

			if result != tc.expectedResult {
				t.Errorf("RESULT: got: %v, expect: %v", result, tc.expectedResult)
			}
		})
	}
}

func TestService_processingResolvedTime(t *testing.T) {
	testCases := []struct {
		name           string
		status         string
		expectedResult *time.Time
	}{
		{
			name:           "status_resolved",
			status:         StatusResolved,
			expectedResult: nil,
		},
		{
			name:           "status_active",
			status:         StatusActive,
			expectedResult: nil,
		},
		{
			name:           "status_empty",
			status:         "",
			expectedResult: nil,
		},
		{
			name:           "status_random",
			status:         "random_status",
			expectedResult: nil,
		},
	}

	s := &Service{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := s.processingResolvedTime(tc.status)

			if tc.status == StatusResolved {
				if result == nil {
					t.Error("TIME: got nil, expect non-nil for resolved status")
				} else {
					diff := time.Since(*result)
					if diff > time.Minute {
						t.Errorf("TIME: seems too old: %v", *result)
					}

					location := result.Location()
					if location.String() != "UTC" {
						t.Errorf("TIME: got location: %v, expect: UTC", location)
					}
				}
			} else {
				if result != nil {
					t.Errorf("TIME: got: %v, expect: nil", result)
				}
			}
		})
	}
}

func TestService_FromDtoToEntitie_Integration(t *testing.T) {
	cfg := config.Config{
		DefaultRadius: 5000,
		MaxRadius:     10000,
	}

	testCases := []struct {
		name           string
		request        *dto.RegistrationIncidentRequest
		expectedEntity *entities.RegistrationIncidentEntitie
		expectedError  string
		checkFields    []string
	}{
		{
			name: "valid_request_active_status",
			request: &dto.RegistrationIncidentRequest{
				Name:           "Test Incident",
				Type:           "emergency",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				Description:    getPtrStr("Test Description"),
				RadiusInMeters: getIntPtr(1000),
				Status:         getPtrStr(StatusActive),
			},
			expectedEntity: &entities.RegistrationIncidentEntitie{
				Name:         "Test Incident",
				Type:         "emergency",
				Latitude:     "55.7558",
				Longitude:    "37.6173",
				Description:  getPtrStr("Test Description"),
				Status:       StatusActive,
				IsActive:     true,
				Radius:       1000,
				ResolvedTime: nil,
			},
			checkFields: []string{"Name", "Type", "Latitude", "Longitude", "Description", "Status", "IsActive", "Radius", "ResolvedTime"},
		},
		{
			name: "valid_request_resolved_status",
			request: &dto.RegistrationIncidentRequest{
				Name:           "Resolved Incident",
				Type:           "maintenance",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				Description:    getPtrStr("Already fixed"),
				RadiusInMeters: getIntPtr(500),
				Status:         getPtrStr(StatusResolved),
			},
			expectedEntity: &entities.RegistrationIncidentEntitie{
				Name:        "Resolved Incident",
				Type:        "maintenance",
				Latitude:    "55.7558",
				Longitude:   "37.6173",
				Description: getPtrStr("Already fixed"),
				Status:      StatusResolved,
				IsActive:    false,
				Radius:      500,
			},
			checkFields: []string{"Name", "Type", "Latitude", "Longitude", "Description", "Status", "IsActive", "Radius"},
		},
		{
			name: "valid_request_without_optional_fields",
			request: &dto.RegistrationIncidentRequest{
				Name:      "Minimal Incident",
				Type:      "alert",
				Latitude:  "55.7558",
				Longitude: "37.6173",
			},
			expectedEntity: &entities.RegistrationIncidentEntitie{
				Name:        "Minimal Incident",
				Type:        "alert",
				Latitude:    "55.7558",
				Longitude:   "37.6173",
				Description: nil,
				Radius:      cfg.DefaultRadius,
			},
			checkFields: []string{"Name", "Type", "Latitude", "Longitude", "Description", "Radius"},
		},
		{
			name: "valid_request_with_empty_description",
			request: &dto.RegistrationIncidentRequest{
				Name:           "Incident with empty description",
				Type:           "info",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				Description:    getPtrStr(""),
				RadiusInMeters: getIntPtr(2000),
				Status:         getPtrStr(StatusActive),
			},
			expectedEntity: &entities.RegistrationIncidentEntitie{
				Name:         "Incident with empty description",
				Type:         "info",
				Latitude:     "55.7558",
				Longitude:    "37.6173",
				Description:  getPtrStr(""),
				Status:       StatusActive,
				IsActive:     true,
				Radius:       2000,
				ResolvedTime: nil,
			},
			checkFields: []string{"Name", "Type", "Latitude", "Longitude", "Description", "Status", "IsActive", "Radius", "ResolvedTime"},
		},

		{
			name: "edge_request_len_name",
			request: &dto.RegistrationIncidentRequest{
				Name:           strings.Repeat("a", 100),
				Type:           "emergency",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				Description:    getPtrStr("Test Description"),
				RadiusInMeters: getIntPtr(1000),
				Status:         getPtrStr(StatusActive),
			},
			expectedEntity: &entities.RegistrationIncidentEntitie{
				Name:         strings.Repeat("a", 100),
				Type:         "emergency",
				Latitude:     "55.7558",
				Longitude:    "37.6173",
				Description:  getPtrStr("Test Description"),
				Status:       StatusActive,
				IsActive:     true,
				Radius:       1000,
				ResolvedTime: nil,
			},
			checkFields: []string{"Name", "Type", "Latitude", "Longitude", "Description", "Status", "IsActive", "Radius", "ResolvedTime"},
		},
		{
			name: "edge_request_len_type",
			request: &dto.RegistrationIncidentRequest{
				Name:           "new",
				Type:           strings.Repeat("a", 100),
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				Description:    getPtrStr("Test Description"),
				RadiusInMeters: getIntPtr(1000),
				Status:         getPtrStr(StatusActive),
			},
			expectedEntity: &entities.RegistrationIncidentEntitie{
				Name:         "new",
				Type:         strings.Repeat("a", 100),
				Latitude:     "55.7558",
				Longitude:    "37.6173",
				Description:  getPtrStr("Test Description"),
				Status:       StatusActive,
				IsActive:     true,
				Radius:       1000,
				ResolvedTime: nil,
			},
			checkFields: []string{"Name", "Type", "Latitude", "Longitude", "Description", "Status", "IsActive", "Radius", "ResolvedTime"},
		},
		{
			name: "invalid_edge_request_len_status",
			request: &dto.RegistrationIncidentRequest{
				Name:           "new",
				Type:           "fire",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				Description:    getPtrStr("Test Description"),
				RadiusInMeters: getIntPtr(1000),
				Status:         getPtrStr(strings.Repeat("a", 20)),
			},
			expectedError: "invalid status",
			checkFields:   []string{},
		},

		{
			name: "invalid_status",
			request: &dto.RegistrationIncidentRequest{
				Name:           "Test Incident",
				Type:           "emergency",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				RadiusInMeters: getIntPtr(1000),
				Status:         getPtrStr("invalid_status"),
			},
			expectedError: "invalid status",
			checkFields:   []string{},
		},
		{
			name: "invalid_radius_zero",
			request: &dto.RegistrationIncidentRequest{
				Name:           "Test Incident",
				Type:           "emergency",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				RadiusInMeters: getIntPtr(0),
				Status:         getPtrStr(StatusActive),
			},
			expectedError: "radius cannot be <= 0",
			checkFields:   []string{},
		},

		{
			name: "invalid_len_name_long",
			request: &dto.RegistrationIncidentRequest{
				Name:           strings.Repeat("a", 101),
				Type:           "emergency",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr(StatusActive),
			},
			expectedError: "very long name",
			checkFields:   []string{},
		},
		{
			name: "invalid_len_type_long",
			request: &dto.RegistrationIncidentRequest{
				Name:           "new",
				Type:           strings.Repeat("a", 101),
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr(StatusActive),
			},
			expectedError: "very long type",
			checkFields:   []string{},
		},
		{
			name: "invalid_len_status_long",
			request: &dto.RegistrationIncidentRequest{
				Name:           "new",
				Type:           "type",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr(strings.Repeat("a", 21)),
			},
			expectedError: "very long status",
			checkFields:   []string{},
		},

		{
			name: "invalid_radius_too_large",
			request: &dto.RegistrationIncidentRequest{
				Name:           "Test Incident",
				Type:           "emergency",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				RadiusInMeters: getIntPtr(15000),
				Status:         getPtrStr(StatusActive),
			},
			expectedError: "radius cannot be >",
			checkFields:   []string{},
		},
		{
			name: "valid_request_max_radius",
			request: &dto.RegistrationIncidentRequest{
				Name:           "Edge case incident",
				Type:           "test",
				Latitude:       "55.7558",
				Longitude:      "37.6173",
				RadiusInMeters: getIntPtr(10000),
				Status:         getPtrStr(StatusActive),
			},
			expectedEntity: &entities.RegistrationIncidentEntitie{
				Name:         "Edge case incident",
				Type:         "test",
				Latitude:     "55.7558",
				Longitude:    "37.6173",
				Radius:       10000,
				Status:       StatusActive,
				IsActive:     true,
				ResolvedTime: nil,
			},
			checkFields: []string{"Radius", "Status", "IsActive"},
		},
	}

	s := &Service{config: &cfg}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entity, err := s.FromDtoToEntitie(tc.request)
			if err != nil {
				if tc.expectedError == "" {
					t.Errorf("ERROR: got unexpected error: %v", err)
					return
				}
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("ERROR: got: %v, expect to contain: %s", err, tc.expectedError)
				}
				return
			} else if tc.expectedError != "" {
				t.Errorf("ERROR: got nil error, expect: %s", tc.expectedError)
				return
			}
			if entity == nil {
				t.Error("ENTITY: got nil entity")
				return
			}

			for _, field := range tc.checkFields {
				switch field {
				case "Name":
					if entity.Name != tc.expectedEntity.Name {
						t.Errorf("NAME: got: %s, expect: %s", entity.Name, tc.expectedEntity.Name)
					}

				case "Type":
					if entity.Type != tc.expectedEntity.Type {
						t.Errorf("TYPE: got: %s, expect: %s", entity.Type, tc.expectedEntity.Type)
					}

				case "Latitude":
					if entity.Latitude != tc.expectedEntity.Latitude {
						t.Errorf("LATITUDE: got: %s, expect: %s", entity.Latitude, tc.expectedEntity.Latitude)
					}

				case "Longitude":
					if entity.Longitude != tc.expectedEntity.Longitude {
						t.Errorf("LONGITUDE: got: %s, expect: %s", entity.Longitude, tc.expectedEntity.Longitude)
					}

				case "Description":
					if tc.expectedEntity.Description == nil {
						if entity.Description != nil {
							t.Errorf("DESCRIPTION: got: %v, expect: nil", entity.Description)
						}
					} else {
						if entity.Description == nil {
							t.Error("DESCRIPTION: got nil, expect non-nil")
						} else if *entity.Description != *tc.expectedEntity.Description {
							t.Errorf("DESCRIPTION: got: %s, expect: %s",
								*entity.Description, *tc.expectedEntity.Description)
						}
					}

				case "Status":
					if entity.Status != tc.expectedEntity.Status {
						t.Errorf("STATUS: got: %s, expect: %s", entity.Status, tc.expectedEntity.Status)
					}

				case "IsActive":
					if entity.IsActive != tc.expectedEntity.IsActive {
						t.Errorf("IS_ACTIVE: got: %v, expect: %v", entity.IsActive, tc.expectedEntity.IsActive)
					}

				case "Radius":
					if entity.Radius != tc.expectedEntity.Radius {
						t.Errorf("RADIUS: got: %d, expect: %d", entity.Radius, tc.expectedEntity.Radius)
					}

				case "ResolvedTime":
					if tc.expectedEntity.ResolvedTime == nil {
						if entity.ResolvedTime != nil {
							t.Errorf("RESOLVED_TIME: got: %v, expect: nil", entity.ResolvedTime)
						}
					} else {
						if entity.ResolvedTime == nil {
							t.Error("RESOLVED_TIME: got nil, expect non-nil")
						} else {
							diff := time.Since(*entity.ResolvedTime)
							if diff > time.Minute {
								t.Errorf("RESOLVED_TIME: seems too old: %v", entity.ResolvedTime)
							}
							if entity.ResolvedTime.Location().String() != "UTC" {
								t.Errorf("RESOLVED_TIME: wrong location: %v", entity.ResolvedTime.Location())
							}
						}
					}
				}
			}
		})
	}
}

func getIntPtr(i int) *int {
	return &i
}

func getBoolPtr(b bool) *bool {
	return &b
}

func getPtrStr(str string) *string {
	return &str
}
