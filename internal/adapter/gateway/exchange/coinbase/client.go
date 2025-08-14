package coinbase

import (
	"context"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	"github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 4

type Client struct{ c *common.Client }

func New() *Client { return &Client{c: common.New("https://api.exchange.coinbase.com")} }

func NewWithBaseURL(base string) *Client {
	return &Client{c: common.New(base)}
}

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "coinbase" }

type product struct {
	ID     string `json:"id"`              // BTC-USD
	Base   string `json:"base_currency"`
	Quote  string `json:"quote_currency"`
	Status string `json:"status"`          // online
}

func (cl *Client) FetchSpot(ctx context.Context) ([]markets.Item, error) {
	var arr []product
	if err := cl.c.GetJSON(ctx, "/products", &arr); err != nil {
		return nil, err
	}
	out := make([]markets.Item, 0, len(arr))
	for _, p := range arr {
		active := strings.EqualFold(p.Status, "online")
		out = append(out, markets.Item{
			ExchangeID: ExchangeID, Type: markets.TypeSpot, Symbol: p.ID,
			Base: strings.ToUpper(p.Base), Quote: strings.ToUpper(p.Quote),
			Active: active,
		})
	}
	return out, nil
}

func (cl *Client) FetchFutures(ctx context.Context) ([]markets.Item, error) {
	// There is no public list of futures on this endpoint.
	return nil, nil
}
