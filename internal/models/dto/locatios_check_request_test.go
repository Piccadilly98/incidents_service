package dto_test

import (
	"fmt"
	"testing"

	"github.com/Piccadilly98/incidents_service/internal/models/dto"
)

func TestLocationCheckRequest_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		req       *dto.LocationCheckRequest
		wantedErr error
	}{
		{
			name: "valid_coordinates",
			req: &dto.LocationCheckRequest{
				UserID:    "user_1",
				Latitude:  "55.7558",
				Longitude: "37.6173",
			},
		},
		{
			name: "valid_not_user_id",
			req: &dto.LocationCheckRequest{
				Latitude:  "60.0",
				Longitude: "30.0",
			},
			wantedErr: fmt.Errorf("user_id cannot be enpty"),
		},
		{
			name: "invalid_latitude_too_high",
			req: &dto.LocationCheckRequest{
				UserID:    "user_1",
				Latitude:  "90.00000001",
				Longitude: "0",
			},
			wantedErr: fmt.Errorf("latitude incorrect compare"),
		},
		{
			name: "invalid_longitude_too_low",
			req: &dto.LocationCheckRequest{
				UserID:    "user_1",
				Latitude:  "0",
				Longitude: "-180.1",
			},
			wantedErr: fmt.Errorf("longitude incorrect compare"),
		},
		{
			name: "invalid_precision",
			req: &dto.LocationCheckRequest{
				UserID:    "user_2",
				Latitude:  "55.755826123",
				Longitude: "37.6173",
			},
			wantedErr: fmt.Errorf("latitide: incorrect format"),
		},
		{
			name: "nan_latitude",
			req: &dto.LocationCheckRequest{
				UserID:    "user_12",
				Latitude:  "NaN",
				Longitude: "0",
			},
			wantedErr: fmt.Errorf("latitude incorrect parse"),
		},
		{
			name: "nan_longitude",
			req: &dto.LocationCheckRequest{
				UserID:    "user_12",
				Latitude:  "0",
				Longitude: "NaN",
			},
			wantedErr: fmt.Errorf("longitude incorrect parse"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if err != nil {
				if tc.wantedErr != nil {
					if tc.wantedErr.Error() != err.Error() {
						t.Errorf("ERROR: got: %s, expect: %s\n", err.Error(), tc.wantedErr.Error())
					}
				} else {
					t.Errorf("unexpected error: %s\n", err.Error())
				}
			} else {
				if tc.wantedErr != nil {
					t.Errorf("ERORR: got: nil, expect: %s\n", tc.wantedErr.Error())
				}
			}
		})
	}
}
