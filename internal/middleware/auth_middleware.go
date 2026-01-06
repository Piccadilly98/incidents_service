package middleware

import (
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/handlers"
)

func CheckMiddleware(validApiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == validApiKey {
				next.ServeHTTP(w, r)
				return
			}
			handlers.ErrorResponse(w, fmt.Errorf("invalid api-key"), http.StatusForbidden)
		})
	}
}
