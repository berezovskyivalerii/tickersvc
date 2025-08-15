-- +goose Up
BEGIN;

-- 1) Удалить дубли для нового PK (exchange_id, symbol, is_futures)
WITH d AS (
  SELECT ctid
  FROM (
    SELECT ctid,
           ROW_NUMBER() OVER (
             PARTITION BY exchange_id, symbol, is_futures
             ORDER BY ctid
           ) AS rn
    FROM incoming_tickers
  ) t
  WHERE rn > 1
)
DELETE FROM incoming_tickers it
USING d
WHERE it.ctid = d.ctid;

-- 2) PK на staging: (exchange_id, symbol, is_futures)
ALTER TABLE incoming_tickers
  DROP CONSTRAINT IF EXISTS incoming_tickers_pkey;

ALTER TABLE incoming_tickers
  ADD CONSTRAINT incoming_tickers_pkey
  PRIMARY KEY (exchange_id, symbol, is_futures);

-- 3) Уникальность в markets по (exchange_id, symbol, mtype)
DROP INDEX IF EXISTS ux_markets_exchange_symbol;
ALTER TABLE markets
  DROP CONSTRAINT IF EXISTS ux_markets_ex_sym_type;

ALTER TABLE markets
  ADD CONSTRAINT ux_markets_ex_sym_type
  UNIQUE (exchange_id, symbol, mtype);

COMMIT;

-- +goose Down
BEGIN;

ALTER TABLE incoming_tickers
  DROP CONSTRAINT IF EXISTS incoming_tickers_pkey;

ALTER TABLE incoming_tickers
  ADD CONSTRAINT incoming_tickers_pkey
  PRIMARY KEY (exchange_id, symbol);

ALTER TABLE markets
  DROP CONSTRAINT IF EXISTS ux_markets_ex_sym_type;

CREATE UNIQUE INDEX IF NOT EXISTS ux_markets_exchange_symbol
  ON markets(exchange_id, symbol);

COMMIT;
