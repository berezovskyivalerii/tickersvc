package robinhood

import (
	"context"

	"github.com/berezovskyivalerii/tickersvc/internal/adapter/gateway/exchange/common"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

const ExchangeID int16 = 7

type Client struct{ c *common.Client }

func New() *Client  { return &Client{c: common.NewWith("https://api.robinhood.com", common.DefaultOptionsFromEnv())} }
func NewWithBaseURL(base string) *Client { return &Client{c: common.NewWith(base, common.DefaultOptionsFromEnv())} }

func (Client) ExchangeID() int16 { return ExchangeID }
func (Client) Name() string      { return "robinhood" }

// Due to geo-restrictions, often 403. Return empty sets
func (cl *Client) FetchSpot(ctx context.Context) ([]dm.Item, error)    { return nil, nil }
func (cl *Client) FetchFutures(ctx context.Context) ([]dm.Item, error) { return nil, nil }
