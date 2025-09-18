package okx

import (
	"context"
	"strconv"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 3

type Client struct{ c *common.Client }

func New() *Client  { return &Client{c: common.NewWith("https://www.okx.com", common.DefaultOptionsFromEnv())} }
func NewWithBaseURL(base string) *Client { return &Client{c: common.NewWith(base, common.DefaultOptionsFromEnv())} }

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "okx" }

type inst struct {
	Data []struct {
		InstID   string `json:"instId"`   // AAA-USDT, AAA-USDT-SWAP
		InstType string `json:"instType"` // SPOT/SWAP
		BaseCcy  string `json:"baseCcy"`
		QuoteCcy string `json:"quoteCcy"`
		CtVal    string `json:"ctVal"` // can be "1", "0.1", ...
		State    string `json:"state"` // live/suspend
	} `json:"data"`
}

func (cl *Client) FetchSpot(ctx context.Context) ([]dm.Item, error) {
	var v inst
	if err := cl.c.GetJSON(ctx, "/api/v5/public/instruments", map[string]string{"instType": "SPOT"}, &v); err != nil {
		return nil, err
	}
	out := make([]dm.Item, 0, len(v.Data))
	for _, it := range v.Data {
		if !strings.EqualFold(it.State, "live") {
			continue
		}
		out = append(out, dm.Item{
			ExchangeID: cl.ExchangeID(),
			Type:       dm.TypeSpot,
			Symbol:     it.InstID,
			Base:       it.BaseCcy,
			Quote:      it.QuoteCcy,
			Active:     true,
		})
	}
	return out, nil
}

func (cl *Client) FetchFutures(ctx context.Context) ([]dm.Item, error) {
	var v inst
	if err := cl.c.GetJSON(ctx, "/api/v5/public/instruments", map[string]string{"instType": "SWAP"}, &v); err != nil {
		return nil, err
	}
	out := make([]dm.Item, 0, len(v.Data))
	for _, it := range v.Data {
		if !strings.EqualFold(it.State, "live") {
			continue
		}
		var cs *int64
		if it.CtVal != "" {
			if n, err := strconv.ParseFloat(it.CtVal, 64); err == nil && n == float64(int64(n)) {
				x := int64(n)
				cs = &x
			}
		}
		out = append(out, dm.Item{
			ExchangeID:  cl.ExchangeID(),
			Type:        dm.TypeFutures,
			Symbol:      it.InstID,
			Base:        it.BaseCcy,
			Quote:       it.QuoteCcy,
			ContractSize: cs,
			Active:      true,
		})
	}
	return out, nil
}
