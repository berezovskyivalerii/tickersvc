package lists

import "context"

// Item — то, что сохраняем в list_items.
// Futures == nil  → в API отдадим "none".
type Item struct {
	Spot    string
	Futures *string
}

type Repo interface {
	// Replace the contents of a list by its ID entirely (atomically).
	ReplaceByListID(ctx context.Context, listID int16, items []Item) (inserted int, err error)
	// The same, but by slug (we'll find the id, lock the list_defs FOR UPDATE line).
	ReplaceBySlug(ctx context.Context, slug string, items []Item) (inserted int, err error)
}
