package dto_test

import (
	"fmt"
	"testing"

	"github.com/Piccadilly98/incidents_service/internal/models/dto"
)

func TestUpdateRequest_Validate(t *testing.T) {
	testCases := []struct {
		name          string
		dto           *dto.UpdateRequest
		expectedError error
	}{
		{
			name: "valid_1",
			dto: &dto.UpdateRequest{
				Name:        getPtrStr("new"),
				Type:        getPtrStr("type"),
				Description: getPtrStr("text"),
				Radius:      getIntPtr(100),
				Status:      getPtrStr("active"),
			},
		},
		{
			name: "valid_partial_only_type",
			dto: &dto.UpdateRequest{
				Type: getPtrStr("new_type"),
			},
		},
		{
			name: "valid_partial_only_name",
			dto: &dto.UpdateRequest{
				Name: getPtrStr("new"),
			},
		},
		{
			name: "valid_partial_only_radius",
			dto: &dto.UpdateRequest{
				Radius: getIntPtr(1000),
			},
		},
		{
			name: "valid_partial_only_status",
			dto: &dto.UpdateRequest{
				Status: getPtrStr("active"),
			},
		},
		{
			name: "valid_partial_only_description",
			dto: &dto.UpdateRequest{
				Description: getPtrStr("Краткое описание"),
			},
		},

		{
			name:          "no_fields_provided",
			dto:           &dto.UpdateRequest{},
			expectedError: fmt.Errorf("no data for update"),
		},
		{
			name: "all_fields_nil",
			dto: &dto.UpdateRequest{
				Name:        nil,
				Type:        nil,
				Description: nil,
				Radius:      nil,
				Status:      nil,
			},
			expectedError: fmt.Errorf("no data for update"),
		},
		{
			name: "empty_name",
			dto: &dto.UpdateRequest{
				Name: getPtrStr(""),
			},
			expectedError: fmt.Errorf("name cannot be empty"),
		},
		{
			name: "empty_type",
			dto: &dto.UpdateRequest{
				Type: getPtrStr(""),
			},
			expectedError: fmt.Errorf("type cannot be empty"),
		},
		{
			name: "empty_description",
			dto: &dto.UpdateRequest{
				Description: getPtrStr(""),
			},
			expectedError: fmt.Errorf("description cannot be empty"),
		},
		{
			name: "empty_status",
			dto: &dto.UpdateRequest{
				Status: getPtrStr(""),
			},
			expectedError: fmt.Errorf("status cannot be empty"),
		},
		{
			name: "zero_radius",
			dto: &dto.UpdateRequest{
				Radius: getIntPtr(0),
			},
			expectedError: fmt.Errorf("radius canot be <= 0"),
		},
		{
			name: "negative_radius",
			dto: &dto.UpdateRequest{
				Radius: getIntPtr(-50),
			},
			expectedError: fmt.Errorf("radius canot be <= 0"),
		},
		{
			name: "multiple_errors_first_one_returned",
			dto: &dto.UpdateRequest{
				Name:   getPtrStr(""),
				Radius: getIntPtr(0),
			},
			expectedError: fmt.Errorf("name cannot be empty"),
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
