package upbit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
)

func TestFetchSpot_OK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"market":"KRW-BTC"},{"market":"BTC-MOG"},{"market":"USDT-ETH"}]`))
	}))
	defer ts.Close()

	cl := &Client{c: common.New(ts.URL)}
	items, err := cl.FetchSpot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("want 3, got %d", len(items))
	}
	if items[1].Base != "MOG" || items[1].Quote != "BTC" {
		t.Fatalf("bad parse: %+v", items[1])
	}
}
