package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base string
	hc   *http.Client
	opt  Options
}

func New(base string) *Client { return NewWith(base, DefaultOptionsFromEnv()) }

func NewWith(base string, opt Options) *Client {
	if opt.Timeout <= 0 {
		opt.Timeout = 8 * time.Second
	}
	if opt.Retries < 0 {
		opt.Retries = 0
	}
	if opt.BackoffMin <= 0 {
		opt.BackoffMin = 200 * time.Millisecond
	}
	if opt.BackoffMax < opt.BackoffMin {
		opt.BackoffMax = 3 * time.Second
	}
	if opt.UserAgent == "" {
		opt.UserAgent = "tickersvc"
	}
	return &Client{
		base: strings.TrimRight(base, "/"),
		hc:   &http.Client{Timeout: opt.Timeout}, // страховка; ниже ещё ctx.WithTimeout
		opt:  opt,
	}
}

func (c *Client) url(path string, q map[string]string) string {
	u := c.base + path
	if len(q) == 0 {
		return u
	}
	values := url.Values{}
	for k, v := range q {
		values.Set(k, v)
	}
	return u + "?" + values.Encode()
}

func (c *Client) GetJSON(ctx context.Context, path string, q map[string]string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(path, q), nil)
	if err != nil {
		return err
	}
	return c.doJSON(req, v)
}

func (c *Client) doJSON(req *http.Request, v any) error {
	req.Header.Set("User-Agent", c.opt.UserAgent)
	req.Header.Set("Accept", "application/json")

	var lastErr error
	for attempt := 0; attempt <= c.opt.Retries; attempt++ {
		// пер-запросный timeout
		ctx, cancel := context.WithTimeout(req.Context(), c.opt.Timeout)
		r2 := req.Clone(ctx)

		resp, err := c.hc.Do(r2)
		if err != nil {
			cancel()
			if attempt < c.opt.Retries && shouldRetry(0, err) {
				time.Sleep(computeBackoff(c.opt.BackoffMin, c.opt.BackoffMax, attempt, ""))
				lastErr = err
				continue
			}
			return err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			ct := resp.Header.Get("Content-Type")
			if !isJSON(ct) && len(body) > 0 && body[0] != '{' && body[0] != '[' {
				// защитимся от HTML/текст
				return fmt.Errorf("unexpected content-type: %s", ct)
			}
			if v == nil || len(body) == 0 {
				return nil
			}
			dec := json.NewDecoder(bytes.NewReader(body))
			dec.UseNumber()
			if err := dec.Decode(v); err != nil {
				return fmt.Errorf("json decode: %w", err)
			}
			return nil
		}

		// retryable?
		if attempt < c.opt.Retries && shouldRetry(resp.StatusCode, nil) {
			back := computeBackoff(c.opt.BackoffMin, c.opt.BackoffMax, attempt, headerRetryAfter(resp.Header))
			time.Sleep(back)
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			continue
		}
		// не ретраим — вернём тело для отладки (обрежем)
		preview := string(body)
		if len(preview) > 256 {
			preview = preview[:256] + "…"
		}
		return fmt.Errorf("http %d: %s", resp.StatusCode, preview)
	}
	return lastErr
}
