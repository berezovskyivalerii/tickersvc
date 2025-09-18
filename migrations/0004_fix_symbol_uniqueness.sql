-- +goose Up
ALTER TABLE incoming_tickers
  DROP CONSTRAINT IF EXISTS incoming_tickers_pkey;
ALTER TABLE incoming_tickers
  ADD PRIMARY KEY (exchange_id, symbol, is_futures);

DROP INDEX IF EXISTS ux_markets_exchange_symbol;
ALTER TABLE markets
  ADD CONSTRAINT ux_markets_ex_sym_type UNIQUE (exchange_id, symbol, mtype);

-- +goose Down
ALTER TABLE incoming_tickers
  DROP CONSTRAINT IF EXISTS incoming_tickers_pkey;
ALTER TABLE incoming_tickers
  ADD PRIMARY KEY (exchange_id, symbol);

ALTER TABLE markets
  DROP CONSTRAINT IF EXISTS ux_markets_ex_sym_type;

CREATE UNIQUE INDEX IF NOT EXISTS ux_markets_exchange_symbol
  ON markets(exchange_id, symbol);
