package postgres

import (
	"context"
	"database/sql"
	"fmt"

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
func (r *MarketsRepo) SyncSnapshot(ctx context.Context, exID int16, items []markets.Item) (added, updated, archived int, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 1) сериализация по бирже
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, int64(exID)); err != nil {
		return 0, 0, 0, fmt.Errorf("advisory lock: %w", err)
	}

	// 2) очистить staging этой биржи
	if _, err = tx.ExecContext(ctx, `DELETE FROM incoming_tickers WHERE exchange_id = $1`, exID); err != nil {
		return 0, 0, 0, fmt.Errorf("clear incoming: %w", err)
	}

	// 3) загрузить снапшот (upsert в staging)
	insStaging := `
		INSERT INTO incoming_tickers
			(exchange_id, symbol, base_asset, quote_asset, is_futures, contract_size, project_tick)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (exchange_id, symbol, is_futures) DO UPDATE SET
			base_asset    = EXCLUDED.base_asset,
			quote_asset   = EXCLUDED.quote_asset,
			contract_size = EXCLUDED.contract_size,
			project_tick  = EXCLUDED.project_tick
	`
	for _, it := range items {
		isFut := it.Type == markets.TypeFutures
		var cs any
		if it.ContractSize != nil {
			cs = *it.ContractSize
		} else {
			cs = nil
		}
		// project_tick — кладём базовый тикер проекта (у нас = Base)
		if _, err = tx.ExecContext(ctx, insStaging,
			exID, it.Symbol, it.Base, it.Quote, isFut, cs, it.Base,
		); err != nil {
			return 0, 0, 0, fmt.Errorf("insert staging: %w", err)
		}
	}

	// 4) обновить существующие в markets
	updSQL := `
	WITH upd AS (
		UPDATE markets m
		SET base_asset    = it.base_asset,
		    quote_asset   = it.quote_asset,
		    contract_size = it.contract_size,
		    is_active     = TRUE,
		    delisted_at   = NULL
		FROM incoming_tickers it
		WHERE m.exchange_id = $1
		  AND it.exchange_id = $1
		  AND m.symbol = it.symbol
		  AND m.mtype = CASE WHEN it.is_futures THEN 'futures'::market_type ELSE 'spot'::market_type END
		RETURNING 1
	)
	SELECT COUNT(*) FROM upd;
	`
	if err = tx.QueryRowContext(ctx, updSQL, exID).Scan(&updated); err != nil {
		return 0, 0, 0, fmt.Errorf("update markets: %w", err)
	}

	// 5) вставить новые в markets
	insSQL := `
	WITH ins AS (
		INSERT INTO markets
			(exchange_id, mtype, symbol, base_asset, quote_asset, contract_size, is_active, listed_at, delisted_at)
		SELECT  $1,
		        CASE WHEN it.is_futures THEN 'futures'::market_type ELSE 'spot'::market_type END,
		        it.symbol, it.base_asset, it.quote_asset, it.contract_size,
		        TRUE, now(), NULL
		FROM incoming_tickers it
		WHERE it.exchange_id = $1
		  AND NOT EXISTS (
		    SELECT 1 FROM markets m
		    WHERE m.exchange_id = $1
		      AND m.symbol = it.symbol
		      AND m.mtype = CASE WHEN it.is_futures THEN 'futures'::market_type ELSE 'spot'::market_type END
		  )
		RETURNING 1
	)
	SELECT COUNT(*) FROM ins;
	`
	if err = tx.QueryRowContext(ctx, insSQL, exID).Scan(&added); err != nil {
		return 0, 0, 0, fmt.Errorf("insert markets: %w", err)
	}

	// 6) архивировать отсутствующие
	archSQL := `
	WITH arch AS (
		UPDATE markets m
		SET is_active = FALSE,
		    delisted_at = now()
		WHERE m.exchange_id = $1
		  AND m.is_active = TRUE
		  AND NOT EXISTS (
		    SELECT 1 FROM incoming_tickers it
		    WHERE it.exchange_id = $1
		      AND it.symbol = m.symbol
		      AND (CASE WHEN it.is_futures THEN 'futures'::market_type ELSE 'spot'::market_type END) = m.mtype
		  )
		RETURNING 1
	)
	SELECT COUNT(*) FROM arch;
	`
	if err = tx.QueryRowContext(ctx, archSQL, exID).Scan(&archived); err != nil {
		return 0, 0, 0, fmt.Errorf("archive markets: %w", err)
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
