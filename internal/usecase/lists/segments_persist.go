package lists

import (
	"context"
	"fmt"
	"slices"
	"strings"

	listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
	"github.com/berezovskyivalerii/tickersvc/internal/config"
)

func (uc *Interactor) RebuildSegments(ctx context.Context, sources ...string) (map[string]int, error) {
	// 1) множества
	quotes := config.LoadQuotes()
	sets, err := BuildSets(ctx, uc.Markets, quotes)
	if err != nil {
		return nil, fmt.Errorf("build sets: %w", err)
	}

	// 2) сегменты
	all := BuildAllSegments(sets) // map[slug][]listsdom.Row

	// 3) фильтр по источникам (по префиксу slug)
	allow := map[string]bool{}
	if len(sources) > 0 {
		for _, s := range sources {
			switch strings.ToLower(strings.TrimSpace(s)) {
			case "binance", "bybit", "okx":
				allow[s] = true
			}
		}
	}
	filtered := make(map[string][]listsdom.Row, len(all))
	for slug, rows := range all {
		if len(allow) == 0 {
			filtered[slug] = rows
			continue
		}
		if strings.HasPrefix(slug, "binance_") && allow["binance"] {
			filtered[slug] = rows
		}
		if strings.HasPrefix(slug, "bybit_") && allow["bybit"] {
			filtered[slug] = rows
		}
		if strings.HasPrefix(slug, "okx_") && allow["okx"] {
			filtered[slug] = rows
		}
	}

	// 4) список slug-ов, которые реально хотим сохранять
	slugs := make([]string, 0, len(filtered))
	for s := range filtered {
		slugs = append(slugs, s)
	}
	slices.Sort(slugs)

	// 5) найдём id списков по slug через DefsRepo
	slug2id, err := uc.Defs.IDsBySlugs(ctx, slugs)
	if err != nil {
		return nil, fmt.Errorf("load list ids: %w", err)
	}

	// 6) сохраняем
	res := make(map[string]int, len(slugs))
	for _, slug := range slugs {
		id, ok := slug2id[slug]
		if !ok {
			// slug ещё не засеян в list_defs — пропускаем молча
			continue
		}
		rowsCompat := FromDomainRows(filtered[slug]) // доменные → локальные Row (Futures string)
		items := RowsToItems(rowsCompat)             // "none" → NULL
		n, err := uc.Lists.ReplaceByListID(ctx, id, items)
		if err != nil {
			return nil, fmt.Errorf("save %s: %w", slug, err)
		}
		res[slug] = n
	}
	return res, nil
}
