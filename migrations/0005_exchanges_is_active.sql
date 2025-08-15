-- +goose Up
ALTER TABLE exchanges
  ADD COLUMN IF NOT EXISTS is_active boolean NOT NULL DEFAULT true;

-- Robinhoos os off by default (a lot of  403)
UPDATE exchanges SET is_active = false WHERE slug = 'robinhood';

-- +goose Down
ALTER TABLE exchanges
  DROP COLUMN IF EXISTS is_active;
