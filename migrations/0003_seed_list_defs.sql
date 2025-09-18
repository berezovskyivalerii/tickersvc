-- +goose Up
-- Upbit (ignore_btc_only=TRUE)
INSERT INTO list_defs (slug, source_exchange, target_exchange, ignore_btc_only, exclude_any_on_target) VALUES
  ('okx_to_upbit',3,5,TRUE,TRUE),
  ('binance_to_upbit',1,5,TRUE,TRUE),
  ('bybit_to_upbit',2,5,TRUE,TRUE),
  ('robinhood_to_upbit',7,5,TRUE,TRUE)
ON CONFLICT (slug) DO NOTHING;

-- Bithumb (ignore_btc_only=TRUE)
INSERT INTO list_defs (slug, source_exchange, target_exchange, ignore_btc_only, exclude_any_on_target) VALUES
  ('okx_to_bithumb',3,6,TRUE,TRUE),
  ('binance_to_bithumb',1,6,TRUE,TRUE),
  ('bybit_to_bithumb',2,6,TRUE,TRUE),
  ('robinhood_to_bithumb',7,6,TRUE,TRUE)
ON CONFLICT (slug) DO NOTHING;

-- Coinbase (ignore_btc_only=FALSE)
INSERT INTO list_defs (slug, source_exchange, target_exchange, ignore_btc_only, exclude_any_on_target) VALUES
  ('okx_to_coinbase',3,4,FALSE,TRUE),
  ('binance_to_coinbase',1,4,FALSE,TRUE),
  ('bybit_to_coinbase',2,4,FALSE,TRUE),
  ('robinhood_to_coinbase',7,4,FALSE,TRUE)
ON CONFLICT (slug) DO NOTHING;

-- +goose Down
DELETE FROM list_defs WHERE slug IN (
  'okx_to_upbit','binance_to_upbit','bybit_to_upbit','robinhood_to_upbit',
  'okx_to_bithumb','binance_to_bithumb','bybit_to_bithumb','robinhood_to_bithumb',
  'okx_to_coinbase','binance_to_coinbase','bybit_to_coinbase','robinhood_to_coinbase'
);
    