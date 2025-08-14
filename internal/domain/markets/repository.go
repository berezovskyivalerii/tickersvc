package markets

import "context"

type Repo interface {
	SyncSnapshot(ctx context.Context, exchangeID int16, items []Item) (added, updated, archived int, err error)
	LoadActiveByExchange(ctx context.Context, exchangeID int16) ([]Item, error)
}
