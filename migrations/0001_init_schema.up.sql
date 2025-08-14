BEGIN;

-- Exchanges
CREATE TABLE IF NOT EXISTS exchanges (
  id            SMALLINT PRIMARY KEY,
  name          TEXT        NOT NULL,
  slug          TEXT        NOT NULL UNIQUE,
  api_base      TEXT,
  has_spot      BOOLEAN     NOT NULL DEFAULT TRUE,
  has_futures   BOOLEAN     NOT NULL DEFAULT FALSE,
  is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Projects (coins)
CREATE TABLE IF NOT EXISTS projects (
  id         BIGSERIAL PRIMARY KEY,
  ticker     TEXT        NOT NULL UNIQUE, -- e.x: MOG
  name       TEXT        NOT NULL,        -- full project name
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tickers (trading pairs)
CREATE TABLE IF NOT EXISTS tickers (
  id            BIGSERIAL PRIMARY KEY,
  exchange_id   SMALLINT    NOT NULL REFERENCES exchanges(id),
  symbol        TEXT        NOT NULL,            -- as on exchange: "MOGUSDT", "1000MOGUSDT", "MOG-USDT", ...
  base_asset    TEXT        NOT NULL,            -- MOG
  quote_asset   TEXT        NOT NULL,            -- USDT
  is_futures    BOOLEAN     NOT NULL DEFAULT FALSE,
  contract_size BIGINT,                           -- NULL for spot; e.x. 1000 for "1000MOGUSDT"
  is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
  listed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  delisted_at   TIMESTAMPTZ,
  project_id    BIGINT      REFERENCES projects(id),
  CONSTRAINT ux_tickers_exchange_symbol UNIQUE (exchange_id, symbol)
);

-- Indexes for quick search
CREATE INDEX IF NOT EXISTS ix_tickers_active_exchange ON tickers (exchange_id) WHERE is_active;
CREATE INDEX IF NOT EXISTS ix_tickers_base  ON tickers (base_asset);
CREATE INDEX IF NOT EXISTS ix_tickers_quote ON tickers (quote_asset);
CREATE INDEX IF NOT EXISTS ix_tickers_project ON tickers (project_id);

-- Temporary (staging) table for sync
CREATE TABLE IF NOT EXISTS incoming_tickers (
  exchange_id   SMALLINT  NOT NULL,
  symbol        TEXT      NOT NULL,
  base_asset    TEXT      NOT NULL,
  quote_asset   TEXT      NOT NULL,
  is_futures    BOOLEAN   NOT NULL,
  contract_size BIGINT,
  project_tick  TEXT      NOT NULL,    -- usually = base_asset
  PRIMARY KEY (exchange_id, symbol)
);

COMMIT;
