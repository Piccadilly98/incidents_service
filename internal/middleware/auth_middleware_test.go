package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	validApiKey := "valid-api-key"
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "success"}`)
	})

	middleware := CheckMiddleware(validApiKey)
	handler := middleware(nextHandler)

	testCases := []struct {
		name           string
		apiKey         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid_api_key_passes_through",
			apiKey:         validApiKey,
			expectedStatus: http.StatusOK,
			expectedBody:   `"status": "success"`,
		},
		{
			name:           "invalid_api_key_returns_403",
			apiKey:         "wrong-key",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "invalid api-key",
		},
		{
			name:           "missing_api_key_returns_403",
			apiKey:         "",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "invalid api-key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.apiKey != "" {
				req.Header.Set(headerAPI, tc.apiKey)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Status code: got %d, want %d", rr.Code, tc.expectedStatus)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tc.expectedBody) {
				t.Errorf("Response body:\nGOT:  %s\nWANT to contain: %s", body, tc.expectedBody)
			}
		})
	}

}
