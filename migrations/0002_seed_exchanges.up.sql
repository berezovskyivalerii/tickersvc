BEGIN;

INSERT INTO exchanges (id, name, slug, api_base, has_spot, has_futures, is_active) VALUES
  (1, 'Binance',   'binance',   'https://api.binance.com',                 TRUE,  TRUE,  TRUE),
  (2, 'Bybit',     'bybit',     'https://api.bybit.com',                   TRUE,  TRUE,  TRUE),
  (3, 'OKX',       'okx',       'https://www.okx.com',                     TRUE,  TRUE,  TRUE),
  (4, 'Coinbase',  'coinbase',  'https://api.exchange.coinbase.com',       TRUE,  FALSE, TRUE),
  (5, 'Upbit',     'upbit',     'https://api.upbit.com',                   TRUE,  FALSE, TRUE),
  (6, 'Bithumb',   'bithumb',   'https://api.bithumb.com',                 TRUE,  FALSE, TRUE),
  (7, 'Robinhood', 'robinhood', 'https://api.robinhood.com',               TRUE,  FALSE, TRUE)
ON CONFLICT (id) DO NOTHING;

COMMIT;
