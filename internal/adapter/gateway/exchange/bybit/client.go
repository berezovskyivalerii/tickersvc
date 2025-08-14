package bybit

import (
	"context"
	"strconv"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	"github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 2

type Client struct{ c *common.Client }

func New() *Client { return &Client{c: common.New("https://api.bybit.com")} }

func NewWithBaseURL(base string) *Client {
	return &Client{c: common.New(base)}
}

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "bybit" }

type instrumentsResp struct {
	RetCode int `json:"retCode"`
	Result  struct {
		Category string `json:"category"`
		List     []struct {
			Symbol        string `json:"symbol"`
			BaseCoin      string `json:"baseCoin"`
			QuoteCoin     string `json:"quoteCoin"`
			Status        string `json:"status"` // Trading
			ContractSize  string `json:"contractSize"`
		} `json:"list"`
	} `json:"result"`
}

func (cl *Client) fetchCategory(ctx context.Context, cat string) ([]markets.Item, error) {
	var v instrumentsResp
	if err := cl.c.GetJSON(ctx, "/v5/market/instruments-info?category="+cat, &v); err != nil {
		return nil, err
	}
	out := make([]markets.Item, 0, len(v.Result.List))
	for _, it := range v.Result.List {
		active := strings.EqualFold(it.Status, "Trading")
		var csz *int64
		if it.ContractSize != "" {
			if f, err := strconv.ParseFloat(it.ContractSize, 64); err == nil && f > 0 {
				x := int64(f)
				if float64(x) == f {
					csz = &x
				}
			}
		}
		mt := markets.TypeSpot
		if cat == "linear" || cat == "inverse" { mt = markets.TypeFutures }
		out = append(out, markets.Item{
			ExchangeID: ExchangeID, Type: mt, Symbol: it.Symbol,
			Base: strings.ToUpper(it.BaseCoin), Quote: strings.ToUpper(it.QuoteCoin),
			ContractSize: csz, Active: active,
		})
	}
	return out, nil
}

func (cl *Client) FetchSpot(ctx context.Context) ([]markets.Item, error) {
	return cl.fetchCategory(ctx, "spot")
}

func (cl *Client) FetchFutures(ctx context.Context) ([]markets.Item, error) {
	lin, err := cl.fetchCategory(ctx, "linear")
	if err != nil { return nil, err }
	inv, err := cl.fetchCategory(ctx, "inverse")
	if err != nil { return lin, nil } // a part of exchanges can have not a inverse
	return append(lin, inv...), nil
}
