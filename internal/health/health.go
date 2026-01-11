package health

import (
	"context"
	"fmt"
	"net/http"

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
		err := check.PingWithCtx(ctx)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %s", check.Name(), err.Error()))
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
