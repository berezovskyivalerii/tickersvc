package upbit

import (
	"context"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	"github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 5

type Client struct{ c *common.Client }

func New() *Client { return &Client{c: common.New("https://api.upbit.com")} }

func NewWithBaseURL(base string) *Client {
	return &Client{c: common.New(base)}
}

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "upbit" }

type market struct {
	Market string `json:"market"` // e.g. "KRW-BTC", "BTC-MOG", "USDT-ETH"
}

func (cl *Client) FetchSpot(ctx context.Context) ([]markets.Item, error) {
	var arr []market
	if err := cl.c.GetJSON(ctx, "/v1/market/all?isDetails=false", &arr); err != nil {
		return nil, err
	}
	out := make([]markets.Item, 0, len(arr))
	for _, m := range arr {
		parts := strings.Split(m.Market, "-")
		if len(parts) != 2 { continue }
		quote, base := parts[0], parts[1]
		out = append(out, markets.Item{
			ExchangeID: cl.ExchangeID(), Type: markets.TypeSpot,
			Symbol: m.Market, Base: base, Quote: quote, Active: true,
		})
	}
	return out, nil
}

func (cl *Client) FetchFutures(ctx context.Context) ([]markets.Item, error) {
	return nil, nil // there is no futures on Upbit
}
