package db

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/entities"
)

const (
	StatusActive   = "active"
	StatusResolved = "resolved"
	StatusArchived = "archived"
)

func TestPostgresRepository_getQueryAndArgsForUpdate(t *testing.T) {
	someTime := time.Now()
	testCases := []struct {
		name          string
		entit         *entities.UpdateIncident
		id            string
		expectedQuery string
		expectedArgs  []any
	}{
		{
			name: "only_name_update",
			entit: &entities.UpdateIncident{
				Name: getPtrStr("New Fire"),
			},
			id:            "123e4567-e89b-12d3-a456-426614174000",
			expectedQuery: "UPDATE incidents SET name=$1, is_active=$2, resolved_date=$3, updated_date=NOW() WHERE id = $4 RETURNING",
			expectedArgs:  []any{"New Fire", false, (*time.Time)(nil), "123e4567-e89b-12d3-a456-426614174000"},
		},
		{
			name: "only_type_update",
			entit: &entities.UpdateIncident{
				Type: getPtrStr("type"),
			},
			id:            "123e4567-e89b-12d3-a456-426614174000",
			expectedQuery: "UPDATE incidents SET type=$1, is_active=$2, resolved_date=$3, updated_date=NOW() WHERE id = $4 RETURNING",
			expectedArgs:  []any{"type", false, (*time.Time)(nil), "123e4567-e89b-12d3-a456-426614174000"},
		},
		{
			name: "full_update_with_resolve",
			entit: &entities.UpdateIncident{
				Name:         getPtrStr("Big Fire"),
				Radius:       getIntPtr(15000),
				Status:       getPtrStr(StatusResolved),
				IsActive:     false,
				ResolvedTime: &someTime,
			},
			id:            "uuid",
			expectedQuery: "UPDATE incidents SET name=$1, radius=$2, status=$3, is_active=$4, resolved_date=$5, updated_date=NOW() WHERE id = $6 RETURNING",
			expectedArgs:  []any{"Big Fire", 15000, StatusResolved, false, &someTime, "uuid"},
		},
		{
			name: "no_optional_fields",
			entit: &entities.UpdateIncident{
				IsActive:     true,
				ResolvedTime: nil,
			},
			id:            "uuid",
			expectedQuery: "UPDATE incidents SET is_active=$1, resolved_date=$2, updated_date=NOW() WHERE id = $3 RETURNING",
			expectedArgs:  []any{true, (*time.Time)(nil), "uuid"},
		},
		{
			name: "only_resolved_time",
			entit: &entities.UpdateIncident{
				ResolvedTime: &someTime,
			},
			id:            "uuid",
			expectedQuery: "UPDATE incidents SET is_active=$1, resolved_date=$2, updated_date=NOW() WHERE id = $3 RETURNING",
			expectedArgs:  []any{false, &someTime, "uuid"},
		},
		{
			name: "only_description_update",
			entit: &entities.UpdateIncident{
				Description: getPtrStr("New detailed report"),
			},
			id:            "test-id",
			expectedQuery: "UPDATE incidents SET description=$1, is_active=$2, resolved_date=$3, updated_date=NOW() WHERE id = $4 RETURNING",
			expectedArgs:  []any{"New detailed report", false, (*time.Time)(nil), "test-id"},
		},
		{
			name: "only_radius_update",
			entit: &entities.UpdateIncident{
				Radius: getIntPtr(8000),
			},
			id:            "test-id",
			expectedQuery: "UPDATE incidents SET radius=$1, is_active=$2, resolved_date=$3, updated_date=NOW() WHERE id = $4 RETURNING",
			expectedArgs:  []any{8000, false, (*time.Time)(nil), "test-id"},
		},
		{
			name: "only_status_update_to_active",
			entit: &entities.UpdateIncident{
				Status:   getPtrStr(StatusActive),
				IsActive: true,
			},
			id:            "test-id",
			expectedQuery: "UPDATE incidents SET status=$1, is_active=$2, resolved_date=$3, updated_date=NOW() WHERE id = $4 RETURNING",
			expectedArgs:  []any{StatusActive, true, (*time.Time)(nil), "test-id"},
		},
		{
			name: "multiple_fields_without_status",
			entit: &entities.UpdateIncident{
				Name:        getPtrStr("Flood"),
				Type:        getPtrStr("natural disaster"),
				Description: getPtrStr("Heavy rain"),
				Radius:      getIntPtr(20000),
			},
			id:            "test-id",
			expectedQuery: "UPDATE incidents SET name=$1, type=$2, description=$3, radius=$4, is_active=$5, resolved_date=$6, updated_date=NOW() WHERE id = $7 RETURNING",
			expectedArgs:  []any{"Flood", "natural disaster", "Heavy rain", 20000, false, (*time.Time)(nil), "test-id"},
		},
		{
			name: "all_fields_with_archived_and_resolved_time",
			entit: &entities.UpdateIncident{
				Name:         getPtrStr("Earthquake"),
				Type:         getPtrStr("seismic"),
				Description:  getPtrStr("Aftershocks continue"),
				Radius:       getIntPtr(30000),
				Status:       getPtrStr(StatusArchived),
				IsActive:     false,
				ResolvedTime: &someTime,
			},
			id:            "test-id",
			expectedQuery: "UPDATE incidents SET name=$1, type=$2, description=$3, radius=$4, status=$5, is_active=$6, resolved_date=$7, updated_date=NOW() WHERE id = $8 RETURNING",
			expectedArgs:  []any{"Earthquake", "seismic", "Aftershocks continue", 30000, StatusArchived, false, &someTime, "test-id"},
		},
		{
			name: "description_from_nil_to_value",
			entit: &entities.UpdateIncident{
				Description: getPtrStr("First report"),
			},
			id:            "test-id",
			expectedQuery: "UPDATE incidents SET description=$1, is_active=$2, resolved_date=$3, updated_date=NOW() WHERE id = $4 RETURNING",
			expectedArgs:  []any{"First report", false, (*time.Time)(nil), "test-id"},
		},
		{
			name: "description_from_value_to_empty",
			entit: &entities.UpdateIncident{
				Description: getPtrStr(""),
			},
			id:            "test-id",
			expectedQuery: "UPDATE incidents SET description=$1, is_active=$2, resolved_date=$3, updated_date=NOW() WHERE id = $4 RETURNING",
			expectedArgs:  []any{"", false, (*time.Time)(nil), "test-id"},
		},
		{
			name: "reactivate_with_nil_resolved_time",
			entit: &entities.UpdateIncident{
				Status:       getPtrStr(StatusActive),
				IsActive:     true,
				ResolvedTime: nil,
			},
			id:            "test-id",
			expectedQuery: "UPDATE incidents SET status=$1, is_active=$2, resolved_date=$3, updated_date=NOW() WHERE id = $4 RETURNING",
			expectedArgs:  []any{StatusActive, true, (*time.Time)(nil), "test-id"},
		},
	}

	pr := &PostgresRepository{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, args := pr.getQueryAndArgsForUpdate(tc.entit, tc.id)

			if !strings.Contains(query, tc.expectedQuery) {
				t.Errorf("Query mismatch:\nGot:  %s\nWant: %s", query, tc.expectedQuery)
			}
			if len(args) != len(tc.expectedArgs) {
				t.Errorf("Args length mismatch: got %d, want %d", len(args), len(tc.expectedArgs))
			}
			for i, val := range args {
				if val != tc.expectedArgs[i] {
					t.Errorf("ARGS: got: %v, expect: %v\n", val, tc.expectedArgs[i])
				}
			}
		})
	}
}

func TestGetQueryAndArgsForPagination(t *testing.T) {
	testCases := []struct {
		name      string
		input     *entities.PaginationIncidents
		wantQuery string
		wantArgs  []any
	}{
		{
			name:      "empty_filters",
			input:     &entities.PaginationIncidents{},
			wantQuery: "SELECT * FROM incidents;",
			wantArgs:  []any{},
		},
		{
			name: "only_limit_and_offset",
			input: &entities.PaginationIncidents{
				Limit:  10,
				Offset: 20,
			},
			wantQuery: "SELECT * FROM incidents LIMIT $1 OFFSET $2;",
			wantArgs:  []any{10, 20},
		},
		{
			name: "only_id",
			input: &entities.PaginationIncidents{
				ID: "123",
			},
			wantQuery: "SELECT * FROM incidents WHERE id=$1;",
			wantArgs:  []any{"123"},
		},
		{
			name: "only_status",
			input: &entities.PaginationIncidents{
				Status: "active",
			},
			wantQuery: "SELECT * FROM incidents WHERE status=$1;",
			wantArgs:  []any{"active"},
		},
		{
			name: "status_and_type",
			input: &entities.PaginationIncidents{
				Status: "active",
				Type:   "fire",
			},
			wantQuery: "SELECT * FROM incidents WHERE status=$1 AND type=$2;",
			wantArgs:  []any{"active", "fire"},
		},
		{
			name: "type_first_then_status",
			input: &entities.PaginationIncidents{
				Type:   "flood",
				Status: "resolved",
			},
			wantQuery: "SELECT * FROM incidents WHERE status=$1 AND type=$2;",
			wantArgs:  []any{"resolved", "flood"},
		},
		{
			name: "skip_id_but_use_status",
			input: &entities.PaginationIncidents{
				ID:     "",
				Status: "pending",
			},
			wantQuery: "SELECT * FROM incidents WHERE status=$1;",
			wantArgs:  []any{"pending"},
		},
		{
			name: "id_empty_but_type_and_name",
			input: &entities.PaginationIncidents{
				ID:   "",
				Type: "accident",
				Name: "Big crash",
			},
			wantQuery: "SELECT * FROM incidents WHERE type=$1 AND name=$2;",
			wantArgs:  []any{"accident", "Big crash"},
		},
		{
			name: "radius_present",
			input: &entities.PaginationIncidents{
				Radius: func(i int) *int { return &i }(5000),
			},
			wantQuery: "SELECT * FROM incidents WHERE radius=$1;",
			wantArgs:  []any{5000},
		},
		{
			name: "radius_nil_not_included",
			input: &entities.PaginationIncidents{
				Radius: nil,
			},
			wantQuery: "SELECT * FROM incidents;",
			wantArgs:  []any{},
		},
		{
			name: "full_combination",
			input: &entities.PaginationIncidents{
				ID:     "abc-123",
				Status: "active",
				Type:   "theft",
				Name:   "Stolen bike",
				Radius: func(i int) *int { return &i }(3000),
				Limit:  25,
				Offset: 50,
			},
			wantQuery: "SELECT * FROM incidents WHERE id=$1 AND status=$2 AND type=$3 AND name=$4 AND radius=$5 LIMIT $6 OFFSET $7;",
			wantArgs:  []any{"abc-123", "active", "theft", "Stolen bike", 3000, 25, 50},
		},
		{
			name: "limit_zero_not_included",
			input: &entities.PaginationIncidents{
				Status: "done",
				Limit:  0,
				Offset: 0,
			},
			wantQuery: "SELECT * FROM incidents WHERE status=$1;",
			wantArgs:  []any{"done"},
		},
		{
			name: "only name",
			input: &entities.PaginationIncidents{
				Name: "new",
			},
			wantQuery: "SELECT * FROM incidents WHERE name=$1;",
			wantArgs:  []any{"new"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pr := &PostgresRepository{}
			gotQuery, gotArgs := pr.getQueryAndArgsForPagination(tc.input)

			if gotQuery != tc.wantQuery {
				t.Errorf("\nQuery mismatch:\nGOT:  %s\nWANT: %s", gotQuery, tc.wantQuery)
			}

			if !reflect.DeepEqual(gotArgs, tc.wantArgs) {
				t.Errorf("\nArgs mismatch:\nGOT:  %v\nWANT: %v", gotArgs, tc.wantArgs)
			}
		})
	}
}

func getIntPtr(i int) *int {
	return &i
}

func getPtrStr(str string) *string {
	return &str
}
