package bithumb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cl "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/bithumb"
)

func newClient(ts *httptest.Server) *cl.Client {
	return cl.NewWithBaseURL(ts.URL)
}

func TestFetch_All_KRW_USDT(t *testing.T) {
	dataKRW := map[string]any{
		"AAA": map[string]string{"opening_price": "1"},
		"date": "123",
	}
	dataUSDT := map[string]any{
		"BBB": map[string]string{"opening_price": "1"},
		"date": "123",
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/public/ticker/ALL", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"status":"0000","data":dataKRW})
	})
	mux.HandleFunc("/public/ticker/ALL_USDT", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"status":"0000","data":dataUSDT})
	})
	ts := httptest.NewServer(mux); defer ts.Close()
	cli := newClient(ts)

	out, err := cli.FetchSpot(context.Background())
	if err != nil { t.Fatal(err) }
	// должно содержать AAA-KRW и BBB-USDT
	found := map[string]bool{}
	for _, it := range out { found[it.Symbol] = true }
	if !found["AAA-KRW"] || !found["BBB-USDT"] {
		t.Fatalf("missing pairs: %+v", found)
	}
}
