package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type ListsQueryRepo struct{ db *sql.DB }

func NewListsQueryRepo(db *sql.DB) *ListsQueryRepo { return &ListsQueryRepo{db: db} }

func (r *ListsQueryRepo) GetTextBySlug(ctx context.Context, slug string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT li.spot_symbol, COALESCE(li.futures_symbol, 'none')
FROM list_items li
JOIN list_defs ld ON ld.id = li.list_id
WHERE ld.slug = $1
ORDER BY li.spot_symbol`, slug)
	if err != nil { return nil, err }
	defer rows.Close()

	var out []string
	for rows.Next() {
		var spot, fut string
		if err := rows.Scan(&spot, &fut); err != nil { return nil, err }
		out = append(out, spot+", "+fut)
	}
	return out, rows.Err()
}

func (r *ListsQueryRepo) GetTextByTarget(ctx context.Context, targetSlug string) (map[string][]string, error) {
    const q = `
        SELECT s.slug as source_slug, li.spot_symbol,
               COALESCE(li.futures_symbol,'none') AS fut
        FROM list_items li
        JOIN list_defs  ld ON ld.id = li.list_id
        JOIN exchanges  s  ON s.id = ld.source_exchange
        JOIN exchanges  t  ON t.id = ld.target_exchange
        WHERE t.slug = $1
        ORDER BY s.slug, li.spot_symbol`
    rows, err := r.db.QueryContext(ctx, q, targetSlug)
    if err != nil {
        return nil, fmt.Errorf("lists.GetTextByTarget: %w", err)
    }
    defer rows.Close()

    out := map[string][]string{}
    for rows.Next() {
        var src, spot, fut string
        if err := rows.Scan(&src, &spot, &fut); err != nil {
            return nil, fmt.Errorf("lists.GetTextByTarget.scan: %w", err)
        }
        out[src] = append(out[src], spot+", "+fut)
    }
    return out, rows.Err()
}

func (r *ListsQueryRepo) GetAllText(ctx context.Context) (map[string]map[string][]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT t.slug AS target_slug, s.slug AS source_slug, li.spot_symbol, COALESCE(li.futures_symbol,'none')
FROM list_items li
JOIN list_defs ld ON ld.id = li.list_id
JOIN exchanges s ON s.id = ld.source_exchange
JOIN exchanges t ON t.id = ld.target_exchange
ORDER BY t.slug, s.slug, li.spot_symbol`)
	if err != nil { return nil, err }
	defer rows.Close()

	out := map[string]map[string][]string{}
	for rows.Next() {
		var tgt, src, spot, fut string
		if err := rows.Scan(&tgt, &src, &spot, &fut); err != nil { return nil, err }
		if out[tgt] == nil { out[tgt] = map[string][]string{} }
		out[tgt][src] = append(out[tgt][src], spot+", "+fut)
	}
	// для стабильности — отсортируем внутри
	for tgt := range out {
		for src := range out[tgt] {
			arr := out[tgt][src]
			sort.Strings(arr)
			out[tgt][src] = arr
		}
	}
	return out, rows.Err()
}

func (r *ListsQueryRepo) GetRowsBySlug(ctx context.Context, slug string) ([]listsdom.Row, error) {
	const q = `
		SELECT li.spot_symbol, li.futures_symbol
		FROM list_items li
		JOIN list_defs ld ON ld.id = li.list_id
		WHERE ld.slug = $1
		ORDER BY li.spot_symbol`
	rows, err := r.db.QueryContext(ctx, q, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []listsdom.Row
	for rows.Next() {
		var spot string
		var fut *string
		if err := rows.Scan(&spot, &fut); err != nil {
			return nil, err
		}
		out = append(out, listsdom.Row{Spot: spot, Futures: fut})
	}
	return out, rows.Err()
}


// (Компилятор требует, чтобы ListsQueryRepo реализовывал интерфейс)
var _ listsdom.QueryRepo = (*ListsQueryRepo)(nil)
