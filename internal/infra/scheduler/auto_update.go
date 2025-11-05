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

	// небольшой рандомный сдвиг старта
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
					// уже идёт автоапдейт — пропускаем тик
					continue
				}

				func() {
					defer atomic.StoreInt32(&a.running, 0)

					cctx, cancel := context.WithTimeout(ctx, timeout)
					defer cancel()

					// 1) синк рынков со всех бирж
					if a.Markets != nil {
						if summary, err := a.Markets.RunAll(cctx); err != nil {
							log.Printf("auto-update: markets sync error: %v", err)
						} else {
							log.Printf("auto-update: markets summary: %+v", summary)
						}
					}

					// 2) пересборка списков + 3) пересборка сегментов
					if uc := a.Lists; uc != nil {
						// обычные списки (binance_to_xxx)
						if updated, err := uc.BuildAndSaveFiltered(cctx, nil, nil); err != nil {
							log.Printf("auto-update: lists rebuild error: %v", err)
						} else {
							log.Printf("auto-update: lists rebuilt (count per slug): %+v", updated)
						}

						// сегменты (binance_seg*, bybit_seg*, okx_seg*)
						if segs, err := uc.RebuildSegments(cctx /* все источники */); err != nil {
							log.Printf("auto-update: segments rebuild error: %v", err)
						} else {
							log.Printf("auto-update: segments rebuilt (count per slug): %+v", segs)
						}
					}
				}()
			}
		}
	}()
}

