package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type ExchangesRepo struct{ db *sql.DB }

func NewExchangesRepo(db *sql.DB) *ExchangesRepo { return &ExchangesRepo{db: db} }

func (r *ExchangesRepo) ActiveMap(ctx context.Context) (map[int16]bool, error) {
	const q = `SELECT id, is_active FROM exchanges`
	m := make(map[int16]bool)
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil { return nil, fmt.Errorf("exchanges active map: %w", err) }
	defer rows.Close()
	for rows.Next() {
		var id int16; var on bool
		if err := rows.Scan(&id, &on); err != nil { return nil, err }
		m[id] = on
	}
	return m, rows.Err()
}

func (r *ExchangesRepo) SetActiveBySlug(ctx context.Context, slug string, on bool) error {
	const q = `UPDATE exchanges SET is_active=$1 WHERE slug=$2`
	_, err := r.db.ExecContext(ctx, q, on, slug)
	if err != nil { return fmt.Errorf("exchanges set active: %w", err) }
	return nil
}
