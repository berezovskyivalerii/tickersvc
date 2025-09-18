package markets

import "context"

type Fetcher interface {
	ExchangeID() int16
	Name() string
	FetchSpot(ctx context.Context) ([]Item, error)
	FetchFutures(ctx context.Context) ([]Item, error) // может вернуть nil,nil если нет фьючей
}
