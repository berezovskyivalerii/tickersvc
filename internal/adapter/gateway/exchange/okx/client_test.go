package okx_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cl "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/okx"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

func newClient(ts *httptest.Server) *cl.Client {
	return cl.NewWithBaseURL(ts.URL)
}

func TestFetch_OKX_SPOT_SWAP(t *testing.T) {
	type payload struct {
		Data []struct {
			InstID, InstType, BaseCcy, QuoteCcy, CtVal, State string
		} `json:"data"`
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v5/public/instruments", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("instType") {
		case "SPOT":
			json.NewEncoder(w).Encode(payload{Data: []struct {
				InstID, InstType, BaseCcy, QuoteCcy, CtVal, State string
			}{
				{"AAA-USDT", "SPOT", "AAA", "USDT", "", "live"},
			}})
		case "SWAP":
			json.NewEncoder(w).Encode(payload{Data: []struct {
				InstID, InstType, BaseCcy, QuoteCcy, CtVal, State string
			}{
				{"AAA-USDT-SWAP", "SWAP", "AAA", "USDT", "1", "live"},
			}})
		default:
			w.WriteHeader(400)
		}
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	cli := newClient(ts)
	ctx := context.Background()

	spot, _ := cli.FetchSpot(ctx)
	fut, _ := cli.FetchFutures(ctx)

	if len(spot) != 1 || spot[0].Type != dm.TypeSpot || spot[0].Base != "AAA" || spot[0].Quote != "USDT" {
		t.Fatalf("bad spot: %+v", spot)
	}
	if len(fut) != 1 || fut[0].Type != dm.TypeFutures || fut[0].Symbol != "AAA-USDT-SWAP" {
		t.Fatalf("bad fut: %+v", fut)
	}
}
