package bithumb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 6

type Client struct{ c *common.Client }

func New() *Client  { return &Client{c: common.NewWith("https://api.bithumb.com", common.DefaultOptionsFromEnv())} }
func NewWithBaseURL(base string) *Client { return &Client{c: common.NewWith(base, common.DefaultOptionsFromEnv())} }

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "bithumb" }

type allResp struct {
	Status string                 `json:"status"` // "0000"
	Data   map[string]json.RawMessage `json:"data"`
}

// /public/ticker/ALL       → KRW-маркет
// /public/ticker/ALL_USDT  → USDT-маркет
func (cl *Client) FetchSpot(ctx context.Context) ([]dm.Item, error) {
	items := make([]dm.Item, 0, 1024)

	parse := func(path, quote string) error {
		var v allResp
		if err := cl.c.GetJSON(ctx, path, nil, &v); err != nil {
			return err
		}
		if v.Status != "0000" {
			return fmt.Errorf("bithumb status=%s", v.Status)
		}
		for k, raw := range v.Data {
			if k == "date" { // служебное поле
				continue
			}
			// если объект, значит монета присутствует
			var tmp map[string]any
			if err := json.Unmarshal(raw, &tmp); err == nil && len(tmp) > 0 {
				items = append(items, dm.Item{
					ExchangeID: cl.ExchangeID(),
					Type:       dm.TypeSpot,
					Symbol:     k + "-" + quote, // "AAA-KRW" / "BBB-USDT"
					Base:       k,
					Quote:      quote,
					Active:     true,
				})
			}
		}
		return nil
	}

	_ = parse("/public/ticker/ALL", "KRW")
	_ = parse("/public/ticker/ALL_USDT", "USDT")

	return items, nil
}

func (cl *Client) FetchFutures(ctx context.Context) ([]dm.Item, error) {
	return nil, nil
}
