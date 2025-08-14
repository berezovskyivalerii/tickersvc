package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/berezovskyivalerii/tickersvc/internal/domain/markets"
)

type MarketsRepo struct {
	db *sql.DB
}

func NewMarketsRepo(db *sql.DB) *MarketsRepo { return &MarketsRepo{db: db} }

// SyncSnapshot atomically synchronizes a snapshot of markets for one exchange:
// 1) loads items into staging (incoming_tickers)
// 2) inserts new ones, updates changed ones/reactivates in markets
// 3) archives missing ones (is_active=false, delisted_at=now())
// returns: added, updated, archived
func (r *MarketsRepo) SyncSnapshot(ctx context.Context, exchangeID int16, items []markets.Item) (int, int, int, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, 0, err
	}
	rollback := func(e error) (int, int, int, error) {
		_ = tx.Rollback()
		return 0, 0, 0, e
	}

	// 1) clear staging for exchange
	if _, err := tx.ExecContext(ctx, `DELETE FROM incoming_tickers WHERE exchange_id = $1`, exchangeID); err != nil {
		return rollback(fmt.Errorf("truncate staging: %w", err))
	}

	// 1.1) bulk insert в staging
	if len(items) > 0 {
		const cols = 7 // exchange_id, symbol, base_asset, quote_asset, is_futures, contract_size, project_tick
		vals := make([]string, 0, len(items))
		args := make([]any, 0, len(items)*cols)
		i := 0
		for _, it := range items {
			i++
			// placeholders: ($1,$2,...$7)
			off := (i-1)*cols + 1
			vals = append(vals, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)", off, off+1, off+2, off+3, off+4, off+5, off+6))

			isFut := it.Type == markets.TypeFutures
			var csize any
			if it.ContractSize != nil {
				csize = *it.ContractSize
			} else {
				csize = nil
			}
			args = append(args,
				exchangeID,
				it.Symbol,
				strings.ToUpper(it.Base),
				strings.ToUpper(it.Quote),
				isFut,
				csize,
				strings.ToUpper(it.Base), // project_tick — используем base как заглушку
			)
		}

		q := `INSERT INTO incoming_tickers
			(exchange_id, symbol, base_asset, quote_asset, is_futures, contract_size, project_tick)
			VALUES ` + strings.Join(vals, ",")
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			return rollback(fmt.Errorf("insert staging: %w", err))
		}
	}

	// 2) upsert from staging into markets (insert new ones)
	var added int
	insertSQL := `
		WITH ins AS (
		INSERT INTO markets (exchange_id, mtype, symbol, base_asset, quote_asset, contract_size, is_active, listed_at)
		SELECT i.exchange_id,
				CASE WHEN i.is_futures THEN 'futures'::market_type ELSE 'spot'::market_type END,
				i.symbol, i.base_asset, i.quote_asset, i.contract_size, TRUE, now()
		FROM incoming_tickers i
		WHERE i.exchange_id = $1
		ON CONFLICT (exchange_id, symbol, mtype) DO NOTHING
		RETURNING 1
		)
		SELECT COALESCE(count(*),0) FROM ins;
	`
	if err := tx.QueryRowContext(ctx, insertSQL, exchangeID).Scan(&added); err != nil {
		return rollback(fmt.Errorf("insert markets: %w", err))
	}

	// 2.1) update changed and reactivation of archived
	var updated int
	updateSQL := `
		WITH upd AS (
		UPDATE markets m
		SET base_asset   = i.base_asset,
			quote_asset  = i.quote_asset,
			contract_size= i.contract_size,
			is_active    = TRUE,
			delisted_at  = NULL,
			mtype        = CASE WHEN i.is_futures THEN 'futures'::market_type ELSE 'spot'::market_type END
		FROM incoming_tickers i
		WHERE m.exchange_id = $1
			AND i.exchange_id = $1
			AND m.symbol = i.symbol
			AND m.mtype  = CASE WHEN i.is_futures THEN 'futures'::market_type ELSE 'spot'::market_type END
			AND (
			m.base_asset IS DISTINCT FROM i.base_asset OR
			m.quote_asset IS DISTINCT FROM i.quote_asset OR
			m.contract_size IS DISTINCT FROM i.contract_size OR
			m.is_active = FALSE
			)
		RETURNING 1
		)
		SELECT COALESCE(count(*),0) FROM upd;
	`
	if err := tx.QueryRowContext(ctx, updateSQL, exchangeID).Scan(&updated); err != nil {
		return rollback(fmt.Errorf("update markets: %w", err))
	}

	// 3) archiving missing characters
	var archived int
	archiveSQL := `
		WITH arc AS (
		UPDATE markets m
		SET is_active = FALSE, delisted_at = now()
		WHERE m.exchange_id = $1
			AND m.is_active = TRUE
			AND NOT EXISTS (
			SELECT 1 FROM incoming_tickers i
			WHERE i.exchange_id = $1
				AND i.symbol = m.symbol
				AND (CASE WHEN i.is_futures THEN 'futures'::market_type ELSE 'spot'::market_type END) = m.mtype
			)
		RETURNING 1
		)
		SELECT COALESCE(count(*),0) FROM arc;
	`
	if err := tx.QueryRowContext(ctx, archiveSQL, exchangeID).Scan(&archived); err != nil {
		return rollback(fmt.Errorf("archive markets: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, err
	}
	return added, updated, archived, nil
}

func (r *MarketsRepo) LoadActiveByExchange(ctx context.Context, exchangeID int16) ([]markets.Item, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT exchange_id, mtype, symbol, base_asset, quote_asset, contract_size, is_active
		FROM markets
		WHERE exchange_id = $1 AND is_active = TRUE
		`, exchangeID)
	if err != nil { return nil, err }
	defer rows.Close()

	var out []markets.Item
	for rows.Next() {
		var it markets.Item
		var mtype string
		var csz *int64
		if err := rows.Scan(&it.ExchangeID, &mtype, &it.Symbol, &it.Base, &it.Quote, &csz, &it.Active); err != nil {
			return nil, err
		}
		if csz != nil { it.ContractSize = csz }
		switch mtype {
		case "spot": it.Type = markets.TypeSpot
		case "futures": it.Type = markets.TypeFutures
		default: it.Type = markets.TypeSpot
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
