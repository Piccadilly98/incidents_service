package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/health"
)

type HealthHandler struct {
	hc *health.HealthChecker
	ew *error_worker.ErrorWorker
}

func NewHealthHandler(hc *health.HealthChecker, ew *error_worker.ErrorWorker) *HealthHandler {
	return &HealthHandler{
		hc: hc,
		ew: ew,
	}
}

func (hh *HealthHandler) Handler(w http.ResponseWriter, r *http.Request) {
	check, code := hh.hc.Check(r.Context())
	if hh.isCanceledCtx(r.Context()) {
		return
	}
	b, err := json.Marshal(check)
	if err != nil {
		processingError(w, err, hh.ew)
		return
	}
	w.Header().Set(HeaderContentType, HeaderJson)
	w.WriteHeader(code)
	w.Write(b)
}

func (hh *HealthHandler) isCanceledCtx(ctx context.Context) bool {
	if err := ctx.Err(); err != nil {
		return true
	}
	return false
}
