package coinbase_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cl "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/coinbase"
)

func newClient(ts *httptest.Server) *cl.Client {
	return cl.NewWithBaseURL(ts.URL)
}

func TestFetch_Products(t *testing.T) {
    type P struct {
        ID            string `json:"id"`
        BaseCurrency  string `json:"base_currency"`
        QuoteCurrency string `json:"quote_currency"`
        Status        string `json:"status"`
    }
    mux := http.NewServeMux()
    mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
        _ = json.NewEncoder(w).Encode([]P{
            {ID: "AAA-USDT", BaseCurrency: "AAA", QuoteCurrency: "USDT", Status: "online"},
            {ID: "BBB-EUR",  BaseCurrency: "BBB", QuoteCurrency: "EUR",  Status: "online"},
        })
    })
    ts := httptest.NewServer(mux)
    defer ts.Close()

    cli := cl.NewWithBaseURL(ts.URL)
    out, err := cli.FetchSpot(context.Background())
    if err != nil { t.Fatal(err) }
    if len(out) != 2 || out[0].Base == "" || out[1].Quote == "" {
        t.Fatalf("bad parse: %+v", out)
    }
}
