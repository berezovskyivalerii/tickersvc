package health

import "context"

type Pinger interface {
	Name() string
	Ping(ctx context.Context) error
}
