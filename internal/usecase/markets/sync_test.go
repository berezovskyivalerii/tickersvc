package marketsuc_test

import (
	"context"
	"testing"
	"time"

	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
	uc "github.com/berezovskyivalerii/tickersvc/internal/usecase/markets"
)

type fakeFetcher struct {
	id   int16
	spot []dm.Item
	fut  []dm.Item
}

func (f fakeFetcher) ExchangeID() int16 { return f.id }
func (f fakeFetcher) Name() string      { return "fake" }
func (f fakeFetcher) FetchSpot(ctx context.Context) ([]dm.Item, error)    { return f.spot, nil }
func (f fakeFetcher) FetchFutures(ctx context.Context) ([]dm.Item, error) { return f.fut, nil }

type fakeRepo struct{ calls int }

func (r *fakeRepo) SyncSnapshot(ctx context.Context, ex int16, items []dm.Item) (int, int, int, error) {
	r.calls++
	return 1, 2, 3, nil
}

func (r *fakeRepo) LoadActiveByExchange(ctx context.Context, exchangeID int16) ([]dm.Item, error){
	r.calls++
	return nil, nil
} 

func TestOrchestrator_RunAll(t *testing.T) {
	f := fakeFetcher{
		id: 1,
		spot: []dm.Item{{ExchangeID: 1, Type: dm.TypeSpot, Symbol: "AAAUSDT", Base: "AAA", Quote: "USDT"}},
		fut:  []dm.Item{{ExchangeID: 1, Type: dm.TypeFutures, Symbol: "AAAUSDT", Base: "AAA", Quote: "USDT"}},
	}
	repo := &fakeRepo{}
	orc := &uc.Orchestrator{Repo: repo, Fetchers: []dm.Fetcher{f}, Timeout: 2 * time.Second}

	sum, err := orc.RunAll(context.Background())
	if err != nil {
		t.Fatalf("RunAll err: %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("repo not called")
	}
	if got := sum[1]; got != [3]int{1, 2, 3} {
		t.Fatalf("summary = %v", got)
	}
}
