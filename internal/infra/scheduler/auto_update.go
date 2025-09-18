package scheduler

import (
	"context"
	"log"
	"math/rand/v2"
	"sync/atomic"
	"time"

	listsuc "github.com/berezovskyivalerii/tickersvc/internal/usecase/lists"
	marketsuc "github.com/berezovskyivalerii/tickersvc/internal/usecase/markets"
)

type AutoUpdater struct {
	Markets *marketsuc.Orchestrator
	Lists   *listsuc.Interactor

	Interval time.Duration
	Timeout  time.Duration

	running int32
}

func (a *AutoUpdater) Start(ctx context.Context) {
	interval := a.Interval
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	timeout := a.Timeout
	if timeout <= 0 {
		timeout = 4 * time.Minute
	}

	time.Sleep(time.Duration(5+rand.IntN(20)) * time.Second)

	t := time.NewTicker(interval)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if !atomic.CompareAndSwapInt32(&a.running, 0, 1) {
					continue
				}
				func() {
					defer atomic.StoreInt32(&a.running, 0)

					cctx, cancel := context.WithTimeout(ctx, timeout)
					defer cancel()

					if a.Markets != nil {
						if summary, err := a.Markets.RunAll(cctx); err != nil {
							log.Printf("auto-update: markets sync error: %v", err)
						} else {
							log.Printf("auto-update: markets summary: %+v", summary)
						}
					}
					if uc := a.Lists; uc != nil {
					updated, err := uc.BuildAndSaveFiltered(cctx, nil, nil)
					if err != nil {
						log.Printf("auto-update: lists rebuild error: %v", err)
					} else {
						log.Printf("auto-update: lists rebuilt (count per slug): %+v", updated)
					}
				}
				}()
			}
		}
	}()
}
