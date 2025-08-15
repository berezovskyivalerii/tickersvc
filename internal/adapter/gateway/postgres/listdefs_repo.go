package postgres

import (
	"context"
	"database/sql"
	"fmt"

	listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type ListDefsRepo struct{ db *sql.DB }

func NewListDefsRepo(db *sql.DB) *ListDefsRepo { return &ListDefsRepo{db: db} }

func (r *ListDefsRepo) Find(ctx context.Context, sourceSlug, targetSlug *string) ([]listsdom.Def, error) {
	q := `
		SELECT ld.id, ld.slug,
			ld.source_exchange, s.slug,
			ld.target_exchange, t.slug
		FROM list_defs ld
		JOIN exchanges s ON s.id = ld.source_exchange AND s.is_active = true
		JOIN exchanges t ON t.id = ld.target_exchange AND t.is_active = true
		WHERE 1=1`
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
	q += " ORDER BY ld.id"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil { return nil, fmt.Errorf("list_defs find: %w", err) }
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


func (r *ListDefsRepo) GetByID(ctx context.Context, id int16) (listsdom.Def, error) {
	const q = `
		SELECT ld.id, ld.slug, ld.source_exchange, s.slug, ld.target_exchange, t.slug
		FROM list_defs ld
		JOIN exchanges s ON s.id = ld.source_exchange
		JOIN exchanges t ON t.id = ld.target_exchange
		WHERE ld.id = $1`
	var d listsdom.Def
	if err := r.db.QueryRowContext(ctx, q, id).
		Scan(&d.ID, &d.Slug, &d.SourceID, &d.SourceSlug, &d.TargetID, &d.TargetSlug); err != nil {
		return listsdom.Def{}, fmt.Errorf("list_defs get by id: %w", err)
	}
	return d, nil
}