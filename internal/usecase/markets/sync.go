package marketsuc

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

type Orchestrator struct {
	Repo     markets.Repo
	Fetchers []markets.Fetcher
	Timeout  time.Duration
	Logger   *slog.Logger
}

func (o *Orchestrator) log() *slog.Logger {
	if o.Logger != nil { return o.Logger }
	return slog.Default()
}

func (o *Orchestrator) RunAll(ctx context.Context) (map[int16][3]int, error) {
	if o.Timeout == 0 { o.Timeout = 30 * time.Second }
	out := make(map[int16][3]int)
	var mu sync.Mutex

	var errsMu sync.Mutex
	var errs []error
	var okMu sync.Mutex
	ok := false

	var wg sync.WaitGroup
	for _, f := range o.Fetchers {
		wg.Add(1)
		f := f
		go func() {
			defer wg.Done()
			l := o.log().With("exchange", f.Name(), "id", f.ExchangeID())
			l.Info("sync start")

			cctx, cancel := context.WithTimeout(ctx, o.Timeout)
			defer cancel()

			spot, errS := f.FetchSpot(cctx)
			fut,  errF := f.FetchFutures(cctx)

			if errS != nil && errF != nil {
				l.Warn("fetch failed", "errSpot", errS, "errFut", errF)
				mu.Lock(); out[f.ExchangeID()] = [3]int{0,0,0}; mu.Unlock()
				errsMu.Lock(); errs = append(errs, errors.New(f.Name()+": "+errS.Error())); errsMu.Unlock()
				return
			}

			items := append(spot, fut...)
			a,u,d,err := o.Repo.SyncSnapshot(cctx, f.ExchangeID(), items)
			if err != nil {
				l.Warn("sync failed", "err", err)
				mu.Lock(); out[f.ExchangeID()] = [3]int{0,0,0}; mu.Unlock()
				errsMu.Lock(); errs = append(errs, errors.New(f.Name()+": "+err.Error())); errsMu.Unlock()
				return
			}
			l.Info("sync done", "added", a, "updated", u, "archived", d)

			mu.Lock(); out[f.ExchangeID()] = [3]int{a,u,d}; mu.Unlock()
			okMu.Lock(); ok = true; okMu.Unlock()
		}()
	}
	wg.Wait()

	// сводка одной строкой в лог
	id2 := map[int16]string{
		1:"binance",2:"bybit",3:"okx",4:"coinbase",5:"upbit",6:"bithumb",7:"robinhood",
	}
	o.log().Info("sync summary\n" + FormatSummary(out, id2))

	if !ok && len(errs) > 0 { return out, errors.Join(errs...) }
	return out, nil
}
