# Схема БД

Ниже — краткая документация под конкретную схему из запроса (`exchanges / markets / list_defs / list_items` + `ENUM market_type`).

---

## `ENUM market_type`

**Значения:** `spot`, `futures`
**Зачем:** типизирует рынок на уровне БД (ENUM), снимает риски «левых» значений.

DDL-фрагмент:

```sql
CREATE TYPE market_type AS ENUM ('spot','futures');
```

---

## `exchanges` — справочник бирж

Единый список бирж с постоянными ID и слагами.

**Поля (ключевое):**

* `id SMALLINT PK` — фиксированный идентификатор биржи
* `name TEXT NOT NULL` — отображаемое название
* `slug TEXT NOT NULL UNIQUE` — машинное имя (e.g. `okx`, `upbit`)
* `is_active BOOLEAN NOT NULL DEFAULT TRUE`
* `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`

**Назначение:** нормализация ссылок из других таблиц.

**Пример:**

```text
id: 1, name: "OKX", slug: "okx", is_active: true, created_at: 2025-08-14
```

---

## `markets` — рынки конкретной биржи

Каждая строка — один инструмент на бирже.

**Поля (ключевое):**

* `id BIGSERIAL PK`
* `exchange_id SMALLINT NOT NULL` → `exchanges(id)`
* `mtype market_type NOT NULL` — `spot`/`futures`
* `symbol TEXT NOT NULL` — как на бирже: `MOGUSDT` / `1000MOGUSDT` / `MOG-USDT`
* `base_asset TEXT NOT NULL`, `quote_asset TEXT NOT NULL`
* `contract_size BIGINT` — для фьючерсов (nullable)
* `is_active BOOLEAN NOT NULL DEFAULT TRUE`
* `listed_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `delisted_at TIMESTAMPTZ`

**Индексы:**

* `ux_markets_exchange_symbol (exchange_id, symbol)` — уникальность инструмента в рамках биржи
* `ix_markets_active` — частичный по `exchange_id` где `is_active = true`
* `ix_markets_base (base_asset)` / `ix_markets_quote (quote_asset)` — быстрый фильтр по активам

**Пример:**

```text
exchange_id: 1, mtype: spot, symbol: "MOGUSDT", base_asset: "MOG", quote_asset: "USDT"
```

---

## `list_defs` — определение списков (источник → цель)

Описывает правила построения списков между биржами.

**Поля (ключевое):**

* `id SMALLSERIAL PK`
* `slug TEXT NOT NULL UNIQUE` — напр.: `okx_to_upbit`
* `source_exchange SMALLINT NOT NULL` → `exchanges(id)`
* `target_exchange SMALLINT NOT NULL` → `exchanges(id)`
* `ignore_btc_only BOOLEAN NOT NULL DEFAULT FALSE` — напр. для Upbit/Bithumb
* `exclude_any_on_target BOOLEAN NOT NULL DEFAULT TRUE` — если уже есть у цели — исключить
* `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`

**Индекс:** `ix_list_defs_src_tgt (source_exchange, target_exchange)`

**Смысл:** декларативное описание трансфера/фильтрации инструментов из источника в цель.

---

## `list_items` — материализованные элементы списка

Фактические элементы, полученные по правилам из `list_defs`.

**Поля (ключевое):**

* `id BIGSERIAL PK`
* `list_id SMALLINT NOT NULL` → `list_defs(id) ON DELETE CASCADE`
* `spot_symbol TEXT NOT NULL` — символ в формате источника (spot)
* `futures_symbol TEXT` — может быть `NULL` → трактуется как «none» в API
* `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`

**Индексы:**

* `ix_list_items_list (list_id)`
* `ux_list_items_unique (list_id, spot_symbol)` — не даём дубли в одном списке

**Назначение:** быстрая отдача готовых списков без пересчёта.

---

### Примеры запросов

**Активные spot-рынки на конкретной бирже:**

```sql
SELECT m.*
FROM markets m
JOIN exchanges e ON e.id = m.exchange_id
WHERE e.slug = 'okx' AND m.mtype = 'spot' AND m.is_active;
```

**Выборка элементов списка с исходной/целевой биржей:**

```sql
SELECT li.spot_symbol, li.futures_symbol, s.slug AS src, t.slug AS tgt
FROM list_items li
JOIN list_defs ld ON ld.id = li.list_id
JOIN exchanges s ON s.id = ld.source_exchange
JOIN exchanges t ON t.id = ld.target_exchange
WHERE ld.slug = 'okx_to_upbit';
```
