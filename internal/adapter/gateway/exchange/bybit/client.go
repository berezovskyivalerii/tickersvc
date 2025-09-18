package bybit

import (
	"context"
	"fmt"
	"strconv"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 2

type Client struct{ c *common.Client }

func New() *Client {
	return &Client{c: common.NewWith("https://api.bybit.com", common.DefaultOptionsFromEnv())}
}
func NewWithBaseURL(base string) *Client {
	return &Client{c: common.NewWith(base, common.DefaultOptionsFromEnv())}
}

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "bybit" }

type instResp struct {
	RetCode int `json:"retCode"`
	Result  struct {
		Category string `json:"category"`
		List     []struct {
			Symbol       string `json:"symbol"`
			BaseCoin     string `json:"baseCoin"`
			QuoteCoin    string `json:"quoteCoin"`
			Status       string `json:"status"` // Trading
			ContractSize string `json:"contractSize"`
		} `json:"list"`
	} `json:"result"`
}

func (cl *Client) FetchSpot(ctx context.Context) ([]dm.Item, error) {
	var out instResp
	if err := cl.c.GetJSON(ctx, "/v5/market/instruments-info", map[string]string{"category": "spot"}, &out); err != nil {
		return nil, err
	}
	if out.RetCode != 0 {
		return nil, fmt.Errorf("bybit retCode=%d", out.RetCode)
	}
	items := make([]dm.Item, 0, len(out.Result.List))
	for _, it := range out.Result.List {
		if it.Status != "Trading" {
			continue
		}
		items = append(items, dm.Item{
			ExchangeID: cl.ExchangeID(),
			Type:       dm.TypeSpot,
			Symbol:     it.Symbol,
			Base:       it.BaseCoin,
			Quote:      it.QuoteCoin,
			Active:     true,
		})
	}
	return items, nil
}

func (cl *Client) FetchFutures(ctx context.Context) ([]dm.Item, error) {
	var out instResp
	if err := cl.c.GetJSON(ctx, "/v5/market/instruments-info", map[string]string{"category": "linear"}, &out); err != nil {
		return nil, err
	}
	if out.RetCode != 0 {
		return nil, fmt.Errorf("bybit retCode=%d", out.RetCode)
	}
	items := make([]dm.Item, 0, len(out.Result.List))
	for _, it := range out.Result.List {
		if it.Status != "Trading" {
			continue
		}
		// parse contractSize if present and integral
		var cs *int64
		if it.ContractSize != "" {
			if f, err := strconv.ParseFloat(it.ContractSize, 64); err == nil {
				if f == float64(int64(f)) {
					v := int64(f)
					cs = &v
				}
			}
		}
		items = append(items, dm.Item{
			ExchangeID:  cl.ExchangeID(),
			Type:        dm.TypeFutures,
			Symbol:      it.Symbol,
			Base:        it.BaseCoin,
			Quote:       it.QuoteCoin,
			ContractSize: cs,
			Active:      true,
		})
	}
	return items, nil
}

