package error_worker

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestErrorWorker_ProcessingError(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		expectedCode int
		expectedErr  error
	}{
		{
			name:         "error nil",
			err:          nil,
			expectedCode: http.StatusOK,
			expectedErr:  nil,
		},

		// ===== БАЗОВЫЕ SQL ОШИБКИ =====
		{
			name:         "sql_no_rows",
			err:          sql.ErrNoRows,
			expectedCode: http.StatusNotFound,
			expectedErr:  fmt.Errorf("not found id"),
		},
		{
			name:         "context_cancel_err_context_package",
			err:          context.Canceled,
			expectedCode: -1,
			expectedErr:  nil,
		},
		{
			name:         "context_cancel_err_custom",
			err:          fmt.Errorf("context canceled"),
			expectedCode: -1,
			expectedErr:  nil,
		},
		{
			name:         "context_cancel_err_custom_upper",
			err:          fmt.Errorf("CONTEXT CANCELED"),
			expectedCode: -1,
			expectedErr:  nil,
		},
		{
			name:         "context_deadline_context_error",
			err:          context.DeadlineExceeded,
			expectedCode: http.StatusGatewayTimeout,
			expectedErr:  fmt.Errorf("request timeout"),
		},
		{
			name:         "context_deadline_context_error_custom",
			err:          fmt.Errorf("deadline exceeded"),
			expectedCode: http.StatusGatewayTimeout,
			expectedErr:  fmt.Errorf("request timeout"),
		},
		{
			name:         "context_deadline_context_error_custom_upper",
			err:          fmt.Errorf("DEADLINE EXCEEDED"),
			expectedCode: http.StatusGatewayTimeout,
			expectedErr:  fmt.Errorf("request timeout"),
		},

		// ===== ОШИБКИ ПОДКЛЮЧЕНИЯ К БД (503 Service Unavailable) =====
		{
			name:         "sql_connections_refused",
			err:          fmt.Errorf("dial tcp [::1]:5432: connect: connection refused"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_connections_refused_pq",
			err:          fmt.Errorf("pq: connection refused"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_connections_refused_connect",
			err:          fmt.Errorf("connect: connection refused"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_no_such_host_dial",
			err:          fmt.Errorf("dial tcp: lookup postgres: no such host"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_no_such_host",
			err:          fmt.Errorf("no such host"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_host_not_found_pq",
			err:          fmt.Errorf("pq: host 'wrong-host' not found"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},

		// ===== ОШИБКИ "DOES NOT EXIST" (503) =====
		{
			name:         "sql_database_does_not_exist",
			err:          fmt.Errorf("database \"wrong_db\" does not exist"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_role_does_not_exist",
			err:          fmt.Errorf("role \"unknown_user\" does not exist"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_pq_database_does_not_exist",
			err:          fmt.Errorf("pq: database does not exist"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_relation_does_not_exist",
			err:          fmt.Errorf("relation \"subscriptions\" does not exist"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},

		// ===== TOO MANY CONNECTIONS (503) =====
		{
			name:         "sql_too_many_clients_pq",
			err:          fmt.Errorf("pq: sorry, too many clients already"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_too_many_connections",
			err:          fmt.Errorf("too many connections"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},

		// ===== SERVER SHUTTING DOWN (503) =====
		{
			name:         "sql_server_shutting_down",
			err:          fmt.Errorf("server is shutting down"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_database_system_shutting_down",
			err:          fmt.Errorf("database system is shutting down"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_pq_database_shutting_down",
			err:          fmt.Errorf("pq: the database system is shutting down"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},

		// ===== SQL SYNTAX ERRORS (503) =====
		{
			name:         "sql_syntax_error_pq",
			err:          fmt.Errorf("pq: syntax error at or near \"SELECT\""),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_syntax_error_near_from",
			err:          fmt.Errorf("pq: syntax error near \"FROM\""),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_invalid_syntax",
			err:          fmt.Errorf("pq: invalid syntax"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},

		// ===== NETWORK UNREACHABLE (503) =====
		{
			name:         "sql_network_unreachable",
			err:          fmt.Errorf("dial tcp 192.168.1.100:5432: network is unreachable"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_network_unreachable_short",
			err:          fmt.Errorf("network is unreachable"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},
		{
			name:         "sql_pq_network_unreachable",
			err:          fmt.Errorf("pq: network unreachable"),
			expectedCode: http.StatusServiceUnavailable,
			expectedErr:  fmt.Errorf("service unavailable"),
		},

		// ===== ПОЛЬЗОВАТЕЛЬСКИЕ ОШИБКИ (400 Bad Request) =====
		{
			name:         "user_EOF_err",
			err:          io.EOF,
			expectedCode: http.StatusBadRequest,
			expectedErr:  fmt.Errorf("body empty"),
		},
		{
			name:         "invalid_page_num",
			err:          fmt.Errorf("invalid page_num"),
			expectedCode: http.StatusBadRequest,
			expectedErr:  fmt.Errorf("invalid page_num"),
		},
		{
			name:         "invalid_status",
			err:          fmt.Errorf("invalid status"),
			expectedCode: http.StatusBadRequest,
			expectedErr:  fmt.Errorf("invalid status"),
		},
		// ===== КОНФЛИКТЫ (409 Conflict) =====
		{
			name:         "incident already archived",
			err:          fmt.Errorf("incident already archived"),
			expectedCode: http.StatusConflict,
			expectedErr:  fmt.Errorf("incident already archived"),
		},
		{
			name:         "unable to update archived incident",
			err:          fmt.Errorf("unable to update archived incident"),
			expectedCode: http.StatusConflict,
			expectedErr:  fmt.Errorf("unable to update archived incident"),
		},
		// ===== NOT FOUND (404 Not Found) =====
		{
			name:         "invalid incident_id",
			err:          fmt.Errorf("invalid incident_id"),
			expectedCode: http.StatusNotFound,
			expectedErr:  fmt.Errorf("invalid incident_id"),
		},
		// ===== SQL USER ERRORS (400 Bad Request) =====
		{
			name:         "sql_violates_foreign_key",
			err:          fmt.Errorf("pq: violates foreign key constraint id_fkey"),
			expectedCode: http.StatusBadRequest,
			expectedErr:  fmt.Errorf("invalid request"),
		},
		{
			name:         "sql_duplicate_key",
			err:          fmt.Errorf("duplicate key value violates unique constraint"),
			expectedCode: http.StatusConflict,
			expectedErr:  fmt.Errorf("id is not unique"),
		},
		{
			name:         "sql_invalid_input_syntax",
			err:          fmt.Errorf("invalid input syntax for type uuid: \"abc\""),
			expectedCode: http.StatusBadRequest,
			expectedErr:  fmt.Errorf("invalid request"),
		},
		{
			name:         "sql_value_too_long",
			err:          fmt.Errorf("value too long for type character varying(255)"),
			expectedCode: http.StatusBadRequest,
			expectedErr:  fmt.Errorf("value too long"),
		},

		// ===== НЕИЗВЕСТНЫЕ ОШИБКИ (500 Internal Server Error) =====
		{
			name:         "unknown_error",
			err:          fmt.Errorf("some random internal error"),
			expectedCode: http.StatusInternalServerError,
			expectedErr:  fmt.Errorf("internal server error"),
		},
		{
			name:         "unknown_error_with_details",
			err:          fmt.Errorf("panic: runtime error: index out of range"),
			expectedCode: http.StatusInternalServerError,
			expectedErr:  fmt.Errorf("internal server error"),
		},
		{
			name:         "empty_error_string",
			err:          fmt.Errorf(""),
			expectedCode: http.StatusInternalServerError,
			expectedErr:  fmt.Errorf("internal server error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ew := NewErrorWorker(true)
			code, err := ew.ProcessError(tc.err)

			if code != tc.expectedCode {
				t.Errorf("CODE: got: %d, expect: %d\n", code, tc.expectedCode)
			}
			if err != nil {
				if err.Error() != tc.expectedErr.Error() {
					t.Errorf("ERROR: got: %s, expect: %s\n", err.Error(), tc.expectedErr.Error())
				}
			} else {
				if tc.expectedErr != nil {
					t.Errorf("ERROR: got: nil, expect: %s\n", tc.expectedErr.Error())
				}
			}
		})
	}
}
