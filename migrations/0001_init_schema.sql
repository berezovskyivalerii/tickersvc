-- +goose Up
CREATE TYPE market_type AS ENUM ('spot','futures');

CREATE TABLE IF NOT EXISTS exchanges (
  id         SMALLINT PRIMARY KEY,
  name       TEXT        NOT NULL,
  slug       TEXT        NOT NULL UNIQUE,
  is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS markets (
  id            BIGSERIAL PRIMARY KEY,
  exchange_id   SMALLINT    NOT NULL REFERENCES exchanges(id),
  mtype         market_type NOT NULL,
  symbol        TEXT        NOT NULL,
  base_asset    TEXT        NOT NULL,
  quote_asset   TEXT        NOT NULL,
  contract_size BIGINT,
  is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
  listed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  delisted_at   TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_markets_exchange_symbol ON markets(exchange_id, symbol);
CREATE INDEX IF NOT EXISTS ix_markets_active ON markets(exchange_id) WHERE is_active;
CREATE INDEX IF NOT EXISTS ix_markets_base   ON markets(base_asset);
CREATE INDEX IF NOT EXISTS ix_markets_quote  ON markets(quote_asset);

CREATE TABLE IF NOT EXISTS list_defs (
  id                    SMALLSERIAL PRIMARY KEY,
  slug                  TEXT        NOT NULL UNIQUE,
  source_exchange       SMALLINT    NOT NULL REFERENCES exchanges(id),
  target_exchange       SMALLINT    NOT NULL REFERENCES exchanges(id),
  ignore_btc_only       BOOLEAN     NOT NULL DEFAULT FALSE,
  exclude_any_on_target BOOLEAN     NOT NULL DEFAULT TRUE,
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS ix_list_defs_src_tgt ON list_defs(source_exchange, target_exchange);

CREATE TABLE IF NOT EXISTS list_items (
  id             BIGSERIAL PRIMARY KEY,
  list_id        SMALLINT    NOT NULL REFERENCES list_defs(id) ON DELETE CASCADE,
  spot_symbol    TEXT        NOT NULL,
  futures_symbol TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS ix_list_items_list ON list_items(list_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_list_items_unique ON list_items(list_id, spot_symbol);

CREATE TABLE IF NOT EXISTS incoming_tickers (
  exchange_id   SMALLINT  NOT NULL,
  symbol        TEXT      NOT NULL,
  base_asset    TEXT      NOT NULL,
  quote_asset   TEXT      NOT NULL,
  is_futures    BOOLEAN   NOT NULL,
  contract_size BIGINT,
  project_tick  TEXT      NOT NULL,
  PRIMARY KEY (exchange_id, symbol)
);

-- +goose Down
DROP TABLE IF EXISTS list_items;
DROP TABLE IF EXISTS list_defs;
DROP TABLE IF EXISTS incoming_tickers;
DROP TABLE IF EXISTS markets;
DROP TABLE IF EXISTS exchanges;
DROP TYPE IF EXISTS market_type;
