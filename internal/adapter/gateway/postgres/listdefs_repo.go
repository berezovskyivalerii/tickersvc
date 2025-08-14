package postgres

import (
	"context"
	"database/sql"

	listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type ListDefsRepo struct{ db *sql.DB }

func NewListDefsRepo(db *sql.DB) *ListDefsRepo { return &ListDefsRepo{db: db} }

func (r *ListDefsRepo) Find(ctx context.Context, sourceSlug, targetSlug *string) ([]listsdom.Def, error) {
	q := `
SELECT ld.id, ld.slug, ld.source_exchange, s.slug, ld.target_exchange, t.slug
FROM list_defs ld
JOIN exchanges s ON s.id = ld.source_exchange
JOIN exchanges t ON t.id = ld.target_exchange
WHERE 1=1
`
	args := []any{}
	if sourceSlug != nil && *sourceSlug != "" {
		q += " AND s.slug = $1"
		args = append(args, *sourceSlug)
	}
	if targetSlug != nil && *targetSlug != "" {
		if len(args) == 0 {
			q += " AND t.slug = $1"
		} else {
			q += " AND t.slug = $2"
		}
		args = append(args, *targetSlug)
	}
	q += " ORDER BY t.slug, s.slug"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var out []listsdom.Def
	for rows.Next() {
		var d listsdom.Def
		if err := rows.Scan(&d.ID, &d.Slug, &d.SourceID, &d.SourceSlug, &d.TargetID, &d.TargetSlug); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
