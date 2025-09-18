package coinbase

import (
	"context"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 4

type Client struct{ c *common.Client }

func New() *Client  { return &Client{c: common.NewWith("https://api.exchange.coinbase.com", common.DefaultOptionsFromEnv())} }
func NewWithBaseURL(base string) *Client { return &Client{c: common.NewWith(base, common.DefaultOptionsFromEnv())} }

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "coinbase" }

type product struct {
	ID            string `json:"id"`             // AAA-USDT
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
	Status        string `json:"status"` // online
}

func (cl *Client) FetchSpot(ctx context.Context) ([]dm.Item, error) {
	var v []product
	if err := cl.c.GetJSON(ctx, "/products", nil, &v); err != nil {
		return nil, err
	}
	out := make([]dm.Item, 0, len(v))
	for _, p := range v {
		if !strings.EqualFold(p.Status, "online") {
			continue
		}
		out = append(out, dm.Item{
			ExchangeID: cl.ExchangeID(),
			Type:       dm.TypeSpot,
			Symbol:     p.ID,
			Base:       p.BaseCurrency,
			Quote:      p.QuoteCurrency,
			Active:     true,
		})
	}
	return out, nil
}

func (cl *Client) FetchFutures(ctx context.Context) ([]dm.Item, error) {
	// Coinbase — only spot
	return nil, nil
}
