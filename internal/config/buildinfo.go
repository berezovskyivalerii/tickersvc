package config

import "time"

// Fill with -ldflags on build.
var (
	AppName   = "tickersvc"
	Version   = "dev"
	Commit    = "none"
	BuildTime = "" // ISO8601
)

type BuildInfo struct {
	AppName   string
	Version   string
	Commit    string
	BuildTime string
	StartedAt time.Time
}

func NewBuildInfo() BuildInfo {
	return BuildInfo{
		AppName:   AppName,
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		StartedAt: time.Now(),
	}
}
