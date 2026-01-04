package dto_test

import (
	"fmt"
	"testing"

	"github.com/Piccadilly98/incidents_service/internal/models/dto"
)

func TestRegistrationIncidentRequest_Validate(t *testing.T) {
	testCases := []struct {
		name          string
		dto           *dto.RegistrationIncidentRequest
		expectedError error
	}{
		{
			name: "valid_1_description",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "incident_1",
				Type:           "fire",
				Latitude:       "38.01321",
				Longitude:      "38.01321",
				Description:    getPtrStr("fire woods"),
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr("active"),
			},
		},
		{
			name: "valid_2_no_description",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "incident_1",
				Type:           "fire",
				Latitude:       "38.01321",
				Longitude:      "38.01321",
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr("active"),
			},
		},
		{
			name: "valid_3_status_random",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "incident_1",
				Type:           "fire",
				Latitude:       "38.01321",
				Longitude:      "38.01321",
				Description:    getPtrStr("fire woods"),
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr("rand"),
			},
		},

		{
			name: "invalid_1_empty_name",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "",
				Type:           "fire",
				Latitude:       "38.01321",
				Longitude:      "38.01321",
				Description:    getPtrStr("fire woods"),
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr("active"),
			},
			expectedError: fmt.Errorf("name cannot be empty"),
		},
		{
			name: "invalid_2_empty_type",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "fire",
				Type:           "",
				Latitude:       "38.01321",
				Longitude:      "38.01321",
				Description:    getPtrStr("fire woods"),
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr("active"),
			},
			expectedError: fmt.Errorf("type cannot be empty"),
		},
		{
			name: "invalid_3_empty_latitude",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "fire",
				Type:           "fire",
				Latitude:       "",
				Longitude:      "38.01321",
				Description:    getPtrStr("fire woods"),
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr("active"),
			},
			expectedError: fmt.Errorf("Latitude cannot be empty"),
		},
		{
			name: "invalid_4_empty_longitude",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "fire",
				Type:           "fire",
				Latitude:       "38.01321",
				Longitude:      "",
				Description:    getPtrStr("fire woods"),
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr("active"),
			},
			expectedError: fmt.Errorf("Longitude cannot be empty"),
		},
		{
			name: "invalid_5_empty_desctiption",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "fire",
				Type:           "fire",
				Latitude:       "38.01321",
				Longitude:      "38.123",
				Description:    getPtrStr(""),
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr("active"),
			},
			expectedError: fmt.Errorf("desctiption cannot be empty"),
		},
		{
			name: "invalid_6_raduis==0",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "fire",
				Type:           "fire",
				Latitude:       "38.01321",
				Longitude:      "38.123",
				Description:    getPtrStr("test"),
				RadiusInMeters: getIntPtr(0),
				Status:         getPtrStr("active"),
			},
			expectedError: fmt.Errorf("radius canot be <= 0"),
		},
		{
			name: "invalid_7_raduis<0",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "fire",
				Type:           "fire",
				Latitude:       "38.01321",
				Longitude:      "38.123",
				Description:    getPtrStr("test"),
				RadiusInMeters: getIntPtr(-1),
				Status:         getPtrStr("active"),
			},
			expectedError: fmt.Errorf("radius canot be <= 0"),
		},
		{
			name: "invalid_8_status_empty",
			dto: &dto.RegistrationIncidentRequest{
				Name:           "fire",
				Type:           "fire",
				Latitude:       "38.01321",
				Longitude:      "38.123",
				Description:    getPtrStr("test"),
				RadiusInMeters: getIntPtr(100),
				Status:         getPtrStr(""),
			},
			expectedError: fmt.Errorf("status cannot be empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dto.Validate()
			if err != nil {
				if tc.expectedError != nil {
					if tc.expectedError.Error() != err.Error() {
						t.Errorf("ERROR: got: %s, expect: %s\n", err.Error(), tc.expectedError.Error())
					}
				} else {
					t.Errorf("unexpected error: %s\n", err.Error())
				}
			}
		})
	}
}

func TestRegistrationIncidentRequest_ValidateCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		latitude  string
		longitude string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid coordinates with 8 decimals",
			latitude:  "48.85884430",
			longitude: "2.29435060",
			wantErr:   false,
		},
		{
			name:      "Valid coordinates with 0 decimals",
			latitude:  "48",
			longitude: "2",
			wantErr:   false,
		},
		{
			name:      "Valid negative coordinates",
			latitude:  "-33.8688197",
			longitude: "-151.209295",
			wantErr:   false,
		},
		{
			name:      "Valid max latitude",
			latitude:  "90.00000000",
			longitude: "0",
			wantErr:   false,
		},
		{
			name:      "Valid min latitude",
			latitude:  "-90.00000000",
			longitude: "0",
			wantErr:   false,
		},
		{
			name:      "Valid max longitude",
			latitude:  "0",
			longitude: "180.00000000",
			wantErr:   false,
		},
		{
			name:      "Valid min longitude",
			latitude:  "0",
			longitude: "-180.00000000",
			wantErr:   false,
		},
		{
			name:      "Valid with comma as decimal separator",
			latitude:  "48,8588443",
			longitude: "2,2943506",
			wantErr:   false,
		},
		{
			name:      "Valid border values",
			latitude:  "89.99999999",
			longitude: "179.99999999",
			wantErr:   false,
		},

		{
			name:      "Latitude too long",
			latitude:  "-90.000000000",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},
		{
			name:      "Longitude too long",
			latitude:  "0",
			longitude: "-180.000000000",
			wantErr:   true,
			errMsg:    "longitude incorrect",
		},

		{
			name:      "Latitude not a number",
			latitude:  "abc",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},
		{
			name:      "Longitude not a number",
			latitude:  "0",
			longitude: "xyz",
			wantErr:   true,
			errMsg:    "longitude incorrect",
		},
		{
			name:      "Empty latitude",
			latitude:  "",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},
		{
			name:      "Empty longitude",
			latitude:  "0",
			longitude: "",
			wantErr:   true,
			errMsg:    "longitude incorrect",
		},
		{
			name:      "Multiple decimal points in latitude",
			latitude:  "48.858.8443",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},

		{
			name:      "Latitude too high",
			latitude:  "90.00000001",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},
		{
			name:      "Latitude too low",
			latitude:  "-90.00000001",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},
		{
			name:      "Longitude too high",
			latitude:  "0",
			longitude: "180.00000001",
			wantErr:   true,
			errMsg:    "longitude incorrect",
		},
		{
			name:      "Longitude too low",
			latitude:  "0",
			longitude: "-180.00000001",
			wantErr:   true,
			errMsg:    "longitude incorrect",
		},
		{
			name:      "Both out of range",
			latitude:  "100",
			longitude: "200",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},

		{
			name:      "Too many decimals in latitude",
			latitude:  "48.123456789",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitide: invalid format",
		},
		{
			name:      "Too many decimals in longitude",
			latitude:  "0",
			longitude: "2.123456789",
			wantErr:   true,
			errMsg:    "longitude: invalid format",
		},
		{
			name:      "Too many decimals in both",
			latitude:  "48.123456789",
			longitude: "2.123456789",
			wantErr:   true,
			errMsg:    "latitide: invalid format",
		},
		{
			name:      "Exactly 9 decimals in longitude",
			latitude:  "0",
			longitude: "12.345678901",
			wantErr:   true,
			errMsg:    "longitude: invalid format",
		},

		{
			name:      "Scientific notation",
			latitude:  "1.23e-4",
			longitude: "0",
		},
		{
			name:      "Only decimal point",
			latitude:  ".",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},
		{
			name:      "Only minus sign",
			latitude:  "-",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},
		{
			name:      "Minus and point",
			latitude:  "-.",
			longitude: "0",
			wantErr:   true,
			errMsg:    "latitude incorrect",
		},
		// EDGE
		{
			name:      "Latitude exact max length (12 chars)",
			latitude:  "-90.0000000",
			longitude: "0",
			wantErr:   false,
		},
		{
			name:      "Latitude one char over max (13 chars)",
			latitude:  "-90.00000000",
			longitude: "0",
			wantErr:   false,
		},

		{
			name:      "Longitude exact max length (13 chars)",
			latitude:  "0",
			longitude: "-180.0000000",
			wantErr:   false,
		},
		{
			name:      "Longitude one char over max (14 chars)",
			latitude:  "0",
			longitude: "-180.00000000",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &dto.RegistrationIncidentRequest{
				Latitude:  tt.latitude,
				Longitude: tt.longitude,
				Name:      "Test Incident",
				Type:      "test",
			}

			err := req.ValidateCoordinates()

			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("validateCoordinates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateCoordinates() error message = %v, should contain %v",
						err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestRegistrationIncidentRequest_Validate_Integration(t *testing.T) {
	tests := []struct {
		name    string
		req     *dto.RegistrationIncidentRequest
		wantErr bool
	}{
		{
			name: "Complete valid request",
			req: &dto.RegistrationIncidentRequest{
				Name:           "Пожар в лесу",
				Type:           "fire",
				Latitude:       "55.755826",
				Longitude:      "37.617300",
				Description:    getPtrStr("Задымление в районе парка"),
				RadiusInMeters: getIntPtr(500),
				Status:         getPtrStr("active"),
			},
			wantErr: false,
		},
		{
			name: "Valid without optional fields",
			req: &dto.RegistrationIncidentRequest{
				Name:      "ДТП",
				Type:      "accident",
				Latitude:  "55.755826",
				Longitude: "37.617300",
			},
			wantErr: false,
		},
		{
			name: "Invalid coordinates in full request",
			req: &dto.RegistrationIncidentRequest{
				Name:        "ДТП",
				Type:        "accident",
				Latitude:    "95.755826",
				Longitude:   "37.617300",
				Description: getPtrStr("Описание"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
