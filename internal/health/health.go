package health

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/dto"
)

const (
	StatusServerOk    = "ok"
	StatusServerError = "Service Unavailable"
)

type HealthChecker struct {
	checks []Checks
}

func NewHealthChecker(checks []Checks) *HealthChecker {
	return &HealthChecker{checks: checks}
}

func (hc *HealthChecker) Check(ctx context.Context) (*dto.HealthCheckResponse, int) {
	res := &dto.HealthCheckResponse{ServerStatus: StatusServerOk}
	code := http.StatusOK
	for _, check := range hc.checks {
		if ctx.Err() != nil {
			res.Errors = append(res.Errors, "health check aborted: "+ctx.Err().Error())
			return res, http.StatusServiceUnavailable
		}
		ctxTime, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		err := check.PingWithCtx(ctxTime)
		cancel()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				res.Errors = append(res.Errors, fmt.Sprintf("%s: ping timeout", check.Name()))
			} else {
				res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", check.Name(), err))
			}
		}
	}
	if len(res.Errors) > 0 {
		res.ServerStatus = StatusServerError
		code = http.StatusServiceUnavailable
	}

	return res, code
}

type Checks interface {
	Name() string
	PingWithCtx(ctx context.Context) error
}
