package upbit

import (
	"context"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 5

type Client struct{ c *common.Client }

func New() *Client  { return &Client{c: common.NewWith("https://api.upbit.com", common.DefaultOptionsFromEnv())} }
func NewWithBaseURL(base string) *Client { return &Client{c: common.NewWith(base, common.DefaultOptionsFromEnv())} }

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "upbit" }

type mkt struct {
	Market string `json:"market"` // e.g. "KRW-AAA", "BTC-AAA", "USDT-AAA"
}

// spot only
func (cl *Client) FetchSpot(ctx context.Context) ([]dm.Item, error) {
	var v []mkt
	if err := cl.c.GetJSON(ctx, "/v1/market/all", map[string]string{"isDetails": "false"}, &v); err != nil {
		return nil, err
	}
	out := make([]dm.Item, 0, len(v))
	for _, m := range v {
		parts := strings.SplitN(m.Market, "-", 2)
		if len(parts) != 2 {
			continue
		}
		quote := parts[0]
		base := parts[1]
		out = append(out, dm.Item{
			ExchangeID: cl.ExchangeID(),
			Type:       dm.TypeSpot,
			Symbol:     m.Market,
			Base:       base,
			Quote:      quote,
			Active:     true,
		})
	}
	return out, nil
}

func (cl *Client) FetchFutures(ctx context.Context) ([]dm.Item, error) {
	return nil, nil
}
