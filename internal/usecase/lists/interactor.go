package lists

import (
	"context"
	"fmt"

	ldef "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

// Interactor — единая точка: построить список и перезаписать его в БД.
type Interactor struct {
	Defs   ldef.DefsRepo
	Markets dm.Repo   // LoadActiveByExchange
	Lists   ldef.Repo // ReplaceByListID / ReplaceBySlug (внутри — транзакция)
}

func modeForTarget(targetSlug string) string {
	switch targetSlug {
	case "upbit":
		return "upbit"
	case "bithumb":
		return "bithumb"
	case "coinbase":
		return "coinbase"
	case "binance":
		return "binance"
	default:
		return "coinbase"
	}
}

// BuildAndSaveByListID — собрать и перезаписать содержимое конкретного списка по его ID.
func (uc *Interactor) BuildAndSaveByListID(ctx context.Context, listID int16) (int, error) {
	def, err := uc.Defs.GetByID(ctx, listID)
	if err != nil {
		return 0, err
	}
	return uc.buildAndSave(ctx, def)
}

// BuildAndSaveBySlug — то же самое, но по slug (найдём id через Find).
func (uc *Interactor) BuildAndSaveBySlug(ctx context.Context, slug string) (int, error) {
	defs, err := uc.Defs.Find(ctx, nil, nil)
	if err != nil {
		return 0, err
	}
	for _, d := range defs {
		if d.Slug == slug {
			return uc.buildAndSave(ctx, d)
		}
	}
	return 0, fmt.Errorf("list slug not found: %s", slug)
}

// BuildAndSaveFiltered — собрать и записать все списки по фильтрам source/target (nil => все).
// Возвращает map[slug]inserted.
func (uc *Interactor) BuildAndSaveFiltered(ctx context.Context, sourceSlug, targetSlug *string) (map[string]int, error) {
	defs, err := uc.Defs.Find(ctx, sourceSlug, targetSlug)
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(defs))
	for _, d := range defs {
		n, err := uc.buildAndSave(ctx, d)
		if err != nil {
			return nil, err
		}
		out[d.Slug] = n
	}
	return out, nil
}

// --- внутреннее: общий путь сборки + запись ---
func (uc *Interactor) buildAndSave(ctx context.Context, def ldef.Def) (int, error) {
	// 1) тянем рынки
	source, err := uc.Markets.LoadActiveByExchange(ctx, def.SourceID)
	if err != nil {
		return 0, fmt.Errorf("load source(%s): %w", def.SourceSlug, err)
	}
	target, err := uc.Markets.LoadActiveByExchange(ctx, def.TargetID)
	if err != nil {
		return 0, fmt.Errorf("load target(%s): %w", def.TargetSlug, err)
	}

	// 2) строим список по правилам
	mode := modeForTarget(def.TargetSlug)
	rows := BuildListRows(source, target, mode)

	// 3) сохраняем транзакционно (репозиторий внутри делает DELETE+INSERT в tx)
	items := RowsToItems(rows)
	n, err := uc.Lists.ReplaceByListID(ctx, def.ID, items)
	if err != nil {
		return 0, err
	}
	return n, nil
}
