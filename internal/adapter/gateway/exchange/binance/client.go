package binance

import (
	"context"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	"github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const (
	ExchangeID int16 = 1
)

type Client struct{ c *common.Client }

func New() *Client { return &Client{c: common.New("https://api.binance.com")} }
func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "binance" }

type exInfo struct {
	Symbols []struct {
		Symbol string `json:"symbol"`
		Status string `json:"status"` // TRADING
		Base   string `json:"baseAsset"`
		Quote  string `json:"quoteAsset"`
	} `json:"symbols"`
}

func (cl *Client) FetchSpot(ctx context.Context) ([]markets.Item, error) {
	var v exInfo
	if err := cl.c.GetJSON(ctx, "/api/v3/exchangeInfo", &v); err != nil {
		return nil, err
	}
	out := make([]markets.Item, 0, len(v.Symbols))
	for _, s := range v.Symbols {
		active := strings.EqualFold(s.Status, "TRADING")
		out = append(out, markets.Item{
			ExchangeID: cl.ExchangeID(), Type: markets.TypeSpot,
			Symbol: s.Symbol, Base: s.Base, Quote: s.Quote,
			Active: active,
		})
	}
	return out, nil
}

// USD-M фьючерсы (perps): https://fapi.binance.com/fapi/v1/exchangeInfo
func (cl *Client) FetchFutures(ctx context.Context) ([]markets.Item, error) {
	type finfo struct {
		Symbols []struct {
			Symbol string `json:"symbol"`
			Status string `json:"status"` // TRADING
			Base   string `json:"baseAsset"`
			Quote  string `json:"quoteAsset"`
		} `json:"symbols"`
	}
	var v finfo
	c := common.New("https://fapi.binance.com")
	if err := c.GetJSON(ctx, "/fapi/v1/exchangeInfo", &v); err != nil {
		return nil, err
	}
	out := make([]markets.Item, 0, len(v.Symbols))
	for _, s := range v.Symbols {
		active := strings.EqualFold(s.Status, "TRADING")
		out = append(out, markets.Item{
			ExchangeID: cl.ExchangeID(), Type: markets.TypeFutures,
			Symbol: s.Symbol, Base: s.Base, Quote: s.Quote,
			Active: active,
		})
	}
	return out, nil
}
