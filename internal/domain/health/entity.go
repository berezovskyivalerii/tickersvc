package health

import "time"

type Status string

const (
	StatusOK       Status = "ok"
	StatusDegraded Status = "degraded"
)

type Check struct {
	Name   string
	Status Status
}

type Report struct {
	Status    Status
	Checks    []Check
	Version   string
	Commit    string
	BuildTime string
	StartedAt time.Time
	Now       time.Time
}
