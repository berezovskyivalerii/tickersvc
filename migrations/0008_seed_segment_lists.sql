-- +goose Up
BEGIN;

-- BINANCE (source=1)
INSERT INTO list_defs (slug, source_exchange, list_kind, segment)
VALUES
  ('binance_seg1', 1, 'segment', 'seg1'),
  ('binance_seg2', 1, 'segment', 'seg2'),
  ('binance_seg3', 1, 'segment', 'seg3'),
  ('binance_seg4', 1, 'segment', 'seg4')
ON CONFLICT (slug) DO NOTHING;

-- BYBIT (source=2)
INSERT INTO list_defs (slug, source_exchange, list_kind, segment)
VALUES
  ('bybit_seg1', 2, 'segment', 'seg1'),
  ('bybit_seg2', 2, 'segment', 'seg2'),
  ('bybit_seg3', 2, 'segment', 'seg3'),
  ('bybit_seg4', 2, 'segment', 'seg4')
ON CONFLICT (slug) DO NOTHING;

-- OKX (source=3)
INSERT INTO list_defs (slug, source_exchange, list_kind, segment)
VALUES
  ('okx_seg1', 3, 'segment', 'seg1'),
  ('okx_seg2', 3, 'segment', 'seg2'),
  ('okx_seg3', 3, 'segment', 'seg3'),
  ('okx_seg4', 3, 'segment', 'seg4')
ON CONFLICT (slug) DO NOTHING;

COMMIT;

-- +goose Down
BEGIN;
DELETE FROM list_defs WHERE slug IN (
  'binance_seg1','binance_seg2','binance_seg3','binance_seg4',
  'bybit_seg1','bybit_seg2','bybit_seg3','bybit_seg4',
  'okx_seg1','okx_seg2','okx_seg3','okx_seg4'
);
COMMIT;
