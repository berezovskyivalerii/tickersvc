package bithumb

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	"github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 6

type Client struct{ c *common.Client }

func New() *Client { return &Client{c: common.New("https://api.bithumb.com")} }

func NewWithBaseURL(base string) *Client {
	return &Client{c: common.New(base)}
}

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "bithumb" }

type allResp struct {
	Status string          `json:"status"` // "0000"
	Data   json.RawMessage `json:"data"`   // object with coins's keys + "date"
}

func (cl *Client) fetchAll(ctx context.Context, quote string) ([]markets.Item, error) {
	var v allResp
	path := "/public/ticker/ALL_" + quote
	if quote == "" { path = "/public/ticker/ALL" } // KRW по умолчанию
	if err := cl.c.GetJSON(ctx, path, &v); err != nil {
		return nil, err
	}
	// data — object: { "BTC": {...}, "ETH": {...}, "date": "..." }
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(v.Data, &m); err != nil {
		return nil, err
	}
	out := make([]markets.Item, 0, len(m))
	for base := range m {
		if strings.EqualFold(base, "date") { continue }
		q := "KRW"
		if quote != "" { q = strings.ToUpper(quote) }
		// Fix symbol as BASE-QUOTE
		out = append(out, markets.Item{
			ExchangeID: ExchangeID, Type: markets.TypeSpot,
			Symbol: strings.ToUpper(base) + "-" + q,
			Base: strings.ToUpper(base), Quote: q, Active: true,
		})
	}
	return out, nil
}

func (cl *Client) FetchSpot(ctx context.Context) ([]markets.Item, error) {
	krw, err := cl.fetchAll(ctx, "")
	if err != nil { return nil, err }
	usdt, _ := cl.fetchAll(ctx, "USDT") // if not - ignore
	btc, _  := cl.fetchAll(ctx, "BTC")
	return append(append(krw, usdt...), btc...), nil
}

func (cl *Client) FetchFutures(ctx context.Context) ([]markets.Item, error) {
	return nil, nil
}
