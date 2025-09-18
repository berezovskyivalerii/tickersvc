package health

import (
	"context"
	"time"

	domain "github.com/berezovskyivalerii/tickersvc/internal/domain/health"
)

type ReadinessInput struct{}

type ReadinessOutput struct {
	Status    domain.Status
	Checks    map[string]domain.Status
	Version   string
	Commit    string
	BuildTime string
	Uptime    time.Duration
	Now       time.Time
}

type Clock interface{ Now() time.Time }

type ReadinessInteractor struct {
	Pingers   []domain.Pinger
	Version   string
	Commit    string
	BuildTime string
	StartedAt time.Time
	Clock     Clock
	Timeout   time.Duration
}

func (uc *ReadinessInteractor) Execute(ctx context.Context, _ ReadinessInput) ReadinessOutput {
	if uc.Timeout <= 0 {
		uc.Timeout = 500 * time.Millisecond
	}
	checks := make(map[string]domain.Status, len(uc.Pingers))
	overall := domain.StatusOK

	for _, p := range uc.Pingers {
		cctx, cancel := context.WithTimeout(ctx, uc.Timeout)
		err := p.Ping(cctx)
		cancel()

		if err != nil {
			checks[p.Name()] = domain.StatusDegraded
			overall = domain.StatusDegraded
		} else {
			checks[p.Name()] = domain.StatusOK
		}
	}

	now := uc.Clock.Now()
	return ReadinessOutput{
		Status:    overall,
		Checks:    checks,
		Version:   uc.Version,
		Commit:    uc.Commit,
		BuildTime: uc.BuildTime,
		Uptime:    now.Sub(uc.StartedAt).Truncate(time.Second),
		Now:       now.UTC(),
	}
}
