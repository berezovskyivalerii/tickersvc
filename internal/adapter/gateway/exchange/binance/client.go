	package binance

	import (
		"context"
		"strings"

		"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
		dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
	)

	const ExchangeID int16 = 1

	type Client struct {
		spot *common.Client
		fut  *common.Client
	}

	func New() *Client {
		opt := common.DefaultOptionsFromEnv()
		return &Client{
			spot: common.NewWith("https://api.binance.com", opt),
			fut:  common.NewWith("https://fapi.binance.com", opt), // USD-M
		}
	}

	// тесты/DI
	func NewWithBaseURL(spotBase, futuresBase string) *Client {
		opt := common.DefaultOptionsFromEnv()
		return &Client{
			spot: common.NewWith(spotBase, opt),
			fut:  common.NewWith(futuresBase, opt),
		}
	}

	func (Client) ExchangeID() int16 { return ExchangeID }
	func (Client) Name() string      { return "binance" }

	type exInfo struct {
		Symbols []struct {
			Symbol string `json:"symbol"`
			Status string `json:"status"` // TRADING
			Base   string `json:"baseAsset"`
			Quote  string `json:"quoteAsset"`
		} `json:"symbols"`
	}

	func (cl *Client) FetchSpot(ctx context.Context) ([]dm.Item, error) {
		var v exInfo
		if err := cl.spot.GetJSON(ctx, "/api/v3/exchangeInfo", nil, &v); err != nil {
			return nil, err
		}
		out := make([]dm.Item, 0, len(v.Symbols))
		for _, s := range v.Symbols {
			base, quote := s.Base, s.Quote
			if base == "" || quote == "" {
				if b, q, ok := common.SplitKnownQuote(s.Symbol, common.CommonQuoteSet); ok {
					base, quote = b, q
				}
			}
			active := strings.EqualFold(s.Status, "TRADING")
			out = append(out, dm.Item{
				ExchangeID: cl.ExchangeID(),
				Type:       dm.TypeSpot,
				Symbol:     s.Symbol,
				Base:       base,
				Quote:      quote,
				Active:     active,
			})
		}
		return out, nil
	}

	// USD-M futures (perps): https://fapi.binance.com/fapi/v1/exchangeInfo
	func (cl *Client) FetchFutures(ctx context.Context) ([]dm.Item, error) {
		type finfo struct {
			Symbols []struct {
				Symbol string `json:"symbol"`
				Status string `json:"status"` // TRADING
				Base   string `json:"baseAsset"`
				Quote  string `json:"quoteAsset"`
			} `json:"symbols"`
		}
		var v finfo
		if err := cl.fut.GetJSON(ctx, "/fapi/v1/exchangeInfo", nil, &v); err != nil {
			return nil, err
		}
		out := make([]dm.Item, 0, len(v.Symbols))
		for _, s := range v.Symbols {
			base, quote := s.Base, s.Quote
			if base == "" || quote == "" {
				if b, q, ok := common.SplitKnownQuote(s.Symbol, common.CommonQuoteSet); ok {
					base, quote = b, q
				}
			}
			active := strings.EqualFold(s.Status, "TRADING")
			out = append(out, dm.Item{
				ExchangeID: cl.ExchangeID(),
				Type:       dm.TypeFutures,
				Symbol:     s.Symbol,
				Base:       base,
				Quote:      quote,
				Active:     active,
			})
		}
		return out, nil
	}
