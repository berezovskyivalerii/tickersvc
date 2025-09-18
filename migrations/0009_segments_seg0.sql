-- +goose Up
BEGIN;

-- В чек-констрейнте разрешаем 'seg0'
ALTER TABLE list_defs
  DROP CONSTRAINT IF EXISTS ck_list_defs_mode;

ALTER TABLE list_defs
  ADD CONSTRAINT ck_list_defs_mode
  CHECK (
    (list_kind = 'target'  AND target_exchange IS NOT NULL AND segment IS NULL)
    OR
    (list_kind = 'segment' AND target_exchange IS NULL
       AND segment IN ('seg0','seg1','seg2','seg3','seg4'))
  );

-- Заводим новый список для Binance
INSERT INTO list_defs (slug, source_exchange, list_kind, segment)
VALUES ('binance_seg0', 1, 'segment', 'seg0')
ON CONFLICT (slug) DO NOTHING;

COMMIT;

-- +goose Down
BEGIN;

DELETE FROM list_defs WHERE slug = 'binance_seg0';

ALTER TABLE list_defs
  DROP CONSTRAINT IF EXISTS ck_list_defs_mode;

ALTER TABLE list_defs
  ADD CONSTRAINT ck_list_defs_mode
  CHECK (
    (list_kind = 'target'  AND target_exchange IS NOT NULL AND segment IS NULL)
    OR
    (list_kind = 'segment' AND target_exchange IS NULL
       AND segment IN ('seg1','seg2','seg3','seg4'))
  );

COMMIT;
