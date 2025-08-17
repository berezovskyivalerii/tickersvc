-- +goose Up
BEGIN;

-- новые колонки
ALTER TABLE list_defs
  ADD COLUMN IF NOT EXISTS list_kind TEXT NOT NULL DEFAULT 'target';
ALTER TABLE list_defs
  ADD COLUMN IF NOT EXISTS segment   TEXT NULL;

-- для сегментов target_exchange может быть NULL
ALTER TABLE list_defs
  ALTER COLUMN target_exchange DROP NOT NULL;

-- допустимые значения list_kind
-- +goose StatementBegin
DO $mig$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'ck_list_defs_list_kind'
  ) THEN
    ALTER TABLE list_defs
      ADD CONSTRAINT ck_list_defs_list_kind
      CHECK (list_kind IN ('target','segment'));
  END IF;
END
$mig$;
-- +goose StatementEnd

-- взаимоисключающие режимы target/segment
-- +goose StatementBegin
DO $mig$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'ck_list_defs_mode'
  ) THEN
    ALTER TABLE list_defs
      ADD CONSTRAINT ck_list_defs_mode
      CHECK (
        (list_kind = 'target'  AND target_exchange IS NOT NULL AND segment IS NULL)
        OR
        (list_kind = 'segment' AND target_exchange IS NULL     AND segment IN ('seg1','seg2','seg3','seg4'))
      );
  END IF;
END
$mig$;
-- +goose StatementEnd

-- индекс для быстрых выборок сегментов
CREATE INDEX IF NOT EXISTS ix_list_defs_kind_seg ON list_defs(list_kind, segment);

COMMIT;

-- +goose Down
BEGIN;

DROP INDEX IF EXISTS ix_list_defs_kind_seg;

-- +goose StatementBegin
DO $mig$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'ck_list_defs_mode'
  ) THEN
    ALTER TABLE list_defs DROP CONSTRAINT ck_list_defs_mode;
  END IF;
END
$mig$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $mig$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'ck_list_defs_list_kind'
  ) THEN
    ALTER TABLE list_defs DROP CONSTRAINT ck_list_defs_list_kind;
  END IF;
END
$mig$;
-- +goose StatementEnd

-- вернуть как было
ALTER TABLE list_defs
  ALTER COLUMN target_exchange SET NOT NULL;

ALTER TABLE list_defs DROP COLUMN IF EXISTS segment;
ALTER TABLE list_defs DROP COLUMN IF EXISTS list_kind;

COMMIT;
