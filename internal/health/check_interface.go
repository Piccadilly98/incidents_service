package health

import "context"

type Checks interface {
	Name() string
	PingWithCtx(ctx context.Context) error
}
