package okx

import (
	"context"
	"strconv"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	"github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 3

type Client struct{ c *common.Client }

func New() *Client { return &Client{c: common.New("https://www.okx.com")} }

func NewWithBaseURL(base string) *Client {
	return &Client{c: common.New(base)}
}

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "okx" }

type instsResp struct {
	Data []struct {
		InstID   string `json:"instId"`   // BTC-USDT
		InstType string `json:"instType" `// SPOT|SWAP
		BaseCcy  string `json:"baseCcy"`
		QuoteCcy string `json:"quoteCcy"`
		CtVal    string `json:"ctVal"`  
		State    string `json:"state"`    // live|suspend
	} `json:"data"`
}

func parseCtVal(s string) *int64 {
	if s == "" { return nil }
	f, err := strconv.ParseFloat(s, 64); if err != nil || f <= 0 { return nil }
	x := int64(f); if float64(x) != f { return nil }
	return &x
}

func (cl *Client) fetch(ctx context.Context, instType string) ([]markets.Item, error) {
	var v instsResp
	path := "/api/v5/public/instruments?instType=" + instType
	if err := cl.c.GetJSON(ctx, path, &v); err != nil {
		return nil, err
	}
	out := make([]markets.Item, 0, len(v.Data))
	for _, d := range v.Data {
		active := strings.EqualFold(d.State, "live")
		mt := markets.TypeSpot
		if strings.EqualFold(d.InstType, "SWAP") { mt = markets.TypeFutures }
		out = append(out, markets.Item{
			ExchangeID: ExchangeID, Type: mt, Symbol: d.InstID,
			Base: strings.ToUpper(d.BaseCcy), Quote: strings.ToUpper(d.QuoteCcy),
			ContractSize: parseCtVal(d.CtVal), Active: active,
		})
	}
	return out, nil
}

func (cl *Client) FetchSpot(ctx context.Context) ([]markets.Item, error)   { return cl.fetch(ctx, "SPOT") }
func (cl *Client) FetchFutures(ctx context.Context) ([]markets.Item, error){ return cl.fetch(ctx, "SWAP") }
