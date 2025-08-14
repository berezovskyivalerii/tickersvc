package lists

import (
	"context"

	ldef "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

type Updater struct {
	Defs   ldef.DefsRepo
	MRepo  dm.Repo         // markets repo (LoadActiveByExchange)
	Saver  Saver           // lists saver (ReplaceBySlug)
}

func modeForTarget(targetSlug string) string {
	switch targetSlug {
	case "upbit": return "upbit"
	case "bithumb": return "bithumb"
	case "coinbase": return "coinbase"
	case "binance": return "binance"
	default: return "coinbase"
	}
}

// Обновить списки по фильтрам (nil => все)
func (u *Updater) Update(ctx context.Context, sourceSlug, targetSlug *string) (map[string]int, error) {
	defs, err := u.Defs.Find(ctx, sourceSlug, targetSlug)
	if err != nil { return nil, err }

	result := map[string]int{} // slug -> inserted
	for _, d := range defs {
		// источнику — spot+futures; цели — тоже весь markets (для binance важны фьючи)
		source, err := u.MRepo.LoadActiveByExchange(ctx, d.SourceID)
		if err != nil { return nil, err }
		target, err := u.MRepo.LoadActiveByExchange(ctx, d.TargetID)
		if err != nil { return nil, err }

		rows := BuildListRows(source, target, modeForTarget(d.TargetSlug))
		inserted, err := u.Saver.ReplaceBySlug(ctx, d.Slug, RowsToItems(rows))
		if err != nil { return nil, err }
		result[d.Slug] = inserted
	}
	return result, nil
}
