package robinhood

import (
	"context"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	"github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 7

type Client struct{ c *common.Client }

func New() *Client { return &Client{c: common.New("https://api.robinhood.com")} }
func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "robinhood" }

type pairsResp struct {
	Results []struct {
		Symbol      string `json:"symbol"`       // e.g. "DOGEUSD"
		Tradability string `json:"tradability"` // "tradable"
		Base        struct {
			Code string `json:"code"` // DOGE
		} `json:"asset_currency"`
		Quote struct {
			Code string `json:"code"` // USD
		} `json:"quote_currency"`
	} `json:"results"`
}

func (cl *Client) FetchSpot(ctx context.Context) ([]markets.Item, error) {
	var v pairsResp
	// Public list of pairs (maybe need auth for some envitoment)
	if err := cl.c.GetJSON(ctx, "/crypto/currency_pairs/", &v); err != nil {
		return nil, err
	}
	out := make([]markets.Item, 0, len(v.Results))
	for _, r := range v.Results {
		active := strings.EqualFold(r.Tradability, "tradable")
		base := strings.ToUpper(r.Base.Code)
		quote := strings.ToUpper(r.Quote.Code)
		symbol := r.Symbol
		if symbol == "" { symbol = base + quote }
		out = append(out, markets.Item{
			ExchangeID: ExchangeID, Type: markets.TypeSpot,
			Symbol: symbol, Base: base, Quote: quote, Active: active,
		})
	}
	return out, nil
}

func (cl *Client) FetchFutures(ctx context.Context) ([]markets.Item, error) { return nil, nil }
