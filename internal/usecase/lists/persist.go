package lists

import (
	"context"

	listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
	dm "github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

// Saver is a minimal repository contract (so as not to drag a specific implementation).
type Saver interface {
	ReplaceBySlug(ctx context.Context, slug string, items []listsdom.Item) (int, error)
}

// RowsToItems - translate "none" → NULL for futures.
func RowsToItems(rows []Row) []listsdom.Item {
	out := make([]listsdom.Item, 0, len(rows))
	for _, r := range rows {
		var f *string
		if r.Futures != "" && r.Futures != "none" {
			f = &r.Futures
		}
		out = append(out, listsdom.Item{Spot: r.Spot, Futures: f})
	}
	return out
}

// RebuildAndSave — build list and save by slug.
// mode: "upbit" | "bithumb" | "coinbase" | "binance" (см. BuildListRows).
func RebuildAndSave(ctx context.Context, saver Saver, slug string, source, target []dm.Item, mode string) (inserted int, err error) {
	rows := BuildListRows(source, target, mode)
	items := RowsToItems(rows)
	return saver.ReplaceBySlug(ctx, slug, items)
}
