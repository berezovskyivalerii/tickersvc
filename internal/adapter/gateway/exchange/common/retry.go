package common

import (
	"errors"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func shouldRetry(status int, err error) bool {
	if err != nil {
		var ne net.Error
		if errors.Is(err, http.ErrHandlerTimeout) {
			return true
		}
		if errors.As(err, &ne) {
			// сетевые временные
			return ne.Timeout() || ne.Temporary()
		}
		// прочие трансп. ошибки (refused/reset)
		return true
	}
	// HTTP статусы
	if status == http.StatusTooManyRequests {
		return true
	}
	if status >= 500 && status <= 599 {
		return true
	}
	return false
}

func computeBackoff(min, max time.Duration, attempt int, retryAfter string) time.Duration {
	// honor Retry-After
	if retryAfter != "" {
		// seconds?
		if sec, err := strconv.Atoi(retryAfter); err == nil && sec >= 0 {
			return time.Duration(sec) * time.Second
		}
		// HTTP date?
		if t, err := http.ParseTime(retryAfter); err == nil {
			d := time.Until(t)
			if d > 0 {
				return d
			}
		}
	}
	// expo + jitter
	if attempt < 0 {
		attempt = 0
	}
	// 2^attempt * min with cap at max
	back := min << attempt
	if back > max {
		back = max
	}
	// jitter 50%
	j := time.Duration(rand.Int63n(int64(back) / 2))
	return back/2 + j
}

func headerRetryAfter(h http.Header) string {
	// prefer standard
	if v := h.Get("Retry-After"); v != "" {
		return v
	}
	// some vendors
	if v := h.Get("X-RateLimit-Reset"); v != "" {
		return v
	}
	return ""
}

func isJSON(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/json")
}
