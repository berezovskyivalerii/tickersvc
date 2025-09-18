package bybit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cl "github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/bybit"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

func newClient(ts *httptest.Server) *cl.Client {
	cli := cl.NewWithBaseURL(ts.URL)
	return cli
}

func TestFetch_Spot_And_Futures(t *testing.T) {
	// минимальные фикстуры Bybit v5 instruments
	type R = struct {
		RetCode int `json:"retCode"`
		Result  struct {
			Category string `json:"category"`
			List     []struct {
				Symbol, BaseCoin, QuoteCoin, Status, ContractSize string
			} `json:"list"`
		} `json:"result"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v5/market/instruments-info", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("category")
		var resp R
		resp.RetCode = 0
		resp.Result.Category = q
		switch q {
		case "spot":
			resp.Result.List = []struct {
				Symbol, BaseCoin, QuoteCoin, Status, ContractSize string
			}{
				{"AAAUSDT", "AAA", "USDT", "Trading", ""},
				{"BBBUSD", "BBB", "USD", "Trading", ""},
			}
		case "linear":
			resp.Result.List = []struct {
				Symbol, BaseCoin, QuoteCoin, Status, ContractSize string
			}{
				{"AAAUSDT", "AAA", "USDT", "Trading", "1"},
			}
		default:
			resp.Result.List = nil
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	cli := newClient(ts)
	ctx := context.Background()

	spot, err := cli.FetchSpot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(spot) != 2 || spot[0].Type != dm.TypeSpot {
		t.Fatalf("spot parsed wrong: %+v", spot)
	}
	if spot[0].Base != "AAA" || spot[0].Quote != "USDT" {
		t.Fatalf("spot base/quote wrong: %+v", spot[0])
	}

	fut, err := cli.FetchFutures(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(fut) != 1 || fut[0].Type != dm.TypeFutures {
		t.Fatalf("futures parsed wrong: %+v", fut)
	}
	if fut[0].ContractSize == nil || *fut[0].ContractSize != 1 {
		t.Fatalf("contract size wrong: %+v", fut[0])
	}
}
