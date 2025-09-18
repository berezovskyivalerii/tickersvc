package common

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	Timeout    time.Duration // per-request
	Retries    int           // extra attempts (0 => no retry)
	BackoffMin time.Duration
	BackoffMax time.Duration
	UserAgent  string
}

func DefaultOptionsFromEnv() Options {
	parseDur := func(k string, d time.Duration) time.Duration {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			if x, err := time.ParseDuration(v); err == nil {
				return x
			}
		}
		return d
	}
	parseInt := func(k string, d int) int {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			if x, err := strconv.Atoi(v); err == nil {
				return x
			}
		}
		return d
	}
	ua := os.Getenv("HTTP_USER_AGENT")
	if ua == "" {
		ua = "tickersvc (+https://github.com/berezovskyivalerii/tickersvc)"
	}
	return Options{
		Timeout:    parseDur("HTTP_TIMEOUT", 8*time.Second),
		Retries:    parseInt("HTTP_RETRIES", 2),
		BackoffMin: parseDur("HTTP_BACKOFF_MIN", 200*time.Millisecond),
		BackoffMax: parseDur("HTTP_BACKOFF_MAX", 3*time.Second),
		UserAgent:  ua,
	}
}
