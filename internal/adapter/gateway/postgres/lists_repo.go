package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	listsdom "github.com/berezovskyivalerii/tickersvc/internal/domain/lists"
)

type ListsRepo struct{ db *sql.DB }

func NewListsRepo(db *sql.DB) *ListsRepo { return &ListsRepo{db: db} }

func (r *ListsRepo) ReplaceBySlug(ctx context.Context, slug string, items []listsdom.Item) (int, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil { return 0, err }
	rollback := func(e error) (int, error) { _ = tx.Rollback(); return 0, e }

	var listID int16
	err = tx.QueryRowContext(ctx, `SELECT id FROM list_defs WHERE slug=$1 FOR UPDATE`, slug).Scan(&listID)
	if err != nil { return rollback(fmt.Errorf("list slug not found: %w", err)) }

	n, err := replaceByIDTx(ctx, tx, listID, items)
	if err != nil { return rollback(err) }

	// отметим время обновления
	if _, err := tx.ExecContext(ctx, `UPDATE list_defs SET updated_at = $2 WHERE id = $1`, listID, time.Now().UTC()); err != nil {
		return rollback(fmt.Errorf("update list_defs.updated_at: %w", err))
	}

	if err := tx.Commit(); err != nil { return 0, err }
	return n, nil
}

func (r *ListsRepo) ReplaceByListID(ctx context.Context, listID int16, items []listsdom.Item) (int, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil { return 0, err }
	rollback := func(e error) (int, error) { _ = tx.Rollback(); return 0, e }

	// залочим строку определения на время обновления (на случай параллельных апдейтов)
	var existed int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM list_defs WHERE id=$1 FOR UPDATE`, listID).Scan(&existed); err != nil {
		return rollback(fmt.Errorf("list id not found: %w", err))
	}

	n, err := replaceByIDTx(ctx, tx, listID, items)
	if err != nil { return rollback(err) }

	if _, err := tx.ExecContext(ctx, `UPDATE list_defs SET updated_at = $2 WHERE id = $1`, listID, time.Now().UTC()); err != nil {
		return rollback(fmt.Errorf("update list_defs.updated_at: %w", err))
	}

	if err := tx.Commit(); err != nil { return 0, err }
	return n, nil
}

// Внутренний помощник: DELETE + bulk INSERT в рамках уже открытой транзакции.
func replaceByIDTx(ctx context.Context, tx *sql.Tx, listID int16, items []listsdom.Item) (int, error) {
	// Удаляем старое содержимое
	if _, err := tx.ExecContext(ctx, `DELETE FROM list_items WHERE list_id = $1`, listID); err != nil {
		return 0, fmt.Errorf("delete old list_items: %w", err)
	}

	// Пустой список — это валидно (оставили пустым)
	if len(items) == 0 {
		return 0, nil
	}

	// Bulk INSERT
	const cols = 3 // (list_id, spot_symbol, futures_symbol)
	vals := make([]string, 0, len(items))
	args := make([]any, 0, len(items)*cols)

	for i, it := range items {
		off := i*cols + 1
		vals = append(vals, fmt.Sprintf("($%d,$%d,$%d)", off, off+1, off+2))
		args = append(args, listID, it.Spot, it.Futures) // nil → NULL
	}

	q := `INSERT INTO list_items (list_id, spot_symbol, futures_symbol) VALUES ` + strings.Join(vals, ",")
	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		return 0, fmt.Errorf("insert list_items: %w", err)
	}
	return len(items), nil
}

// ReplaceItems перезаписывает все элементы списка.
func (r *ListsRepo) ReplaceItems(ctx context.Context, listID int16, rows []listsdom.Row) (int, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil { return 0, err }
	rollback := func(e error) (int, error) { _ = tx.Rollback(); return 0, e }

	// Лочим строку list_defs, как в ReplaceByListID
	var existed int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM list_defs WHERE id=$1 FOR UPDATE`, listID).Scan(&existed); err != nil {
		return rollback(fmt.Errorf("list id not found: %w", err))
	}

	// Чистим старые элементы
	if _, err := tx.ExecContext(ctx, `DELETE FROM list_items WHERE list_id = $1`, listID); err != nil {
		return rollback(fmt.Errorf("delete old list_items: %w", err))
	}

	// Дедуп по spot_symbol (на всякий)
	type rec struct{ spot string; fut *string }
	uniq := make(map[string]rec, len(rows))
	for _, rr := range rows {
		// если будут дубли — возьмём первый
		if _, ok := uniq[rr.Spot]; !ok {
			uniq[rr.Spot] = rec{spot: rr.Spot, fut: rr.Futures}
		}
	}
	if len(uniq) == 0 {
		// обновим метку времени и выходим
		if _, err := tx.ExecContext(ctx, `UPDATE list_defs SET updated_at = $2 WHERE id = $1`, listID, time.Now().UTC()); err != nil {
			return rollback(fmt.Errorf("update list_defs.updated_at: %w", err))
		}
		if err := tx.Commit(); err != nil { return 0, err }
		return 0, nil
	}

	// Готовим bulk insert
	const cols = 3
	vals := make([]string, 0, len(uniq))
	args := make([]any, 0, len(uniq)*cols)
	i := 0
	for _, v := range uniq {
		off := i*cols + 1
		vals = append(vals, fmt.Sprintf("($%d,$%d,$%d)", off, off+1, off+2))
		args = append(args, listID, v.spot, v.fut) // nil -> NULL
		i++
	}
	q := `INSERT INTO list_items (list_id, spot_symbol, futures_symbol) VALUES ` + strings.Join(vals, ",")

	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		return rollback(fmt.Errorf("insert list_items: %w", err))
	}

	// Обновим updated_at
	if _, err := tx.ExecContext(ctx, `UPDATE list_defs SET updated_at = $2 WHERE id = $1`, listID, time.Now().UTC()); err != nil {
		return rollback(fmt.Errorf("update list_defs.updated_at: %w", err))
	}

	if err := tx.Commit(); err != nil { return 0, err }
	return len(uniq), nil
}

func (r *ListsRepo) GetRowsBySlug(ctx context.Context, slug string) ([]listsdom.Row, error) {
	return nil, nil
}


