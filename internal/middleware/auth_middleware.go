package middleware

import (
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/handlers"
)

const headerAPI = "X-API-Key"

func CheckMiddleware(validApiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(headerAPI)
			if apiKey == validApiKey {
				next.ServeHTTP(w, r)
				return
			}
			handlers.ErrorResponse(w, fmt.Errorf("invalid api-key"), http.StatusForbidden)
		})
	}
}
