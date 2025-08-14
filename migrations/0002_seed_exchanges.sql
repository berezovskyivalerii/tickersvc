-- +goose Up
INSERT INTO exchanges (id, name, slug, is_active) VALUES
  (1,'Binance','binance',TRUE),
  (2,'Bybit','bybit',TRUE),
  (3,'OKX','okx',TRUE),
  (4,'Coinbase','coinbase',TRUE),
  (5,'Upbit','upbit',TRUE),
  (6,'Bithumb','bithumb',TRUE),
  (7,'Robinhood','robinhood',TRUE)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM exchanges WHERE id IN (1,2,3,4,5,6,7);
