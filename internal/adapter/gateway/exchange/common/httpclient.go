package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type Client struct {
	Base string
	HC   *http.Client
}

func New(base string) *Client {
	return &Client{
		Base: base,
		HC: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
				MaxIdleConns: 100, IdleConnTimeout: 90 * time.Second,
			},
		},
	}
}

func (c *Client) GetJSON(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Base+path, nil)
	if err != nil { return err }
	res, err := c.HC.Do(req)
	if err != nil { return err }
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("http %d: %s", res.StatusCode, string(b))
	}
	return json.NewDecoder(res.Body).Decode(v)
}
