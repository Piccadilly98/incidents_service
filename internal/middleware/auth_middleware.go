package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/handlers"
)

func CheckMiddleware(validApiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == validApiKey {
				ctxValue := context.WithValue(r.Context(), handlers.ContextKeyValidApiKey, true)
				next.ServeHTTP(w, r.WithContext(ctxValue))
				return
			}
			handlers.ErrorResponse(w, fmt.Errorf("invalid api-key"), http.StatusForbidden)

		})
	}
}
