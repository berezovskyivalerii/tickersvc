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

Причина простая: ключ читается из окружения **при старте процесса**. Ты экспортировал `ADMIN_API_KEY` уже после запуска сервиса — у запущенного процесса окружение не меняется. Поэтому мидлварь честно говорит `admin endpoint restricted`.

Сделай одно из двух (или оба):

1. запусти сервис с ключом в окружении;
2. или разреши доступ с локальной машины через белый список CIDR.

Ниже — как правильно и короткая дока.

---

# Админ-доступ к `/admin/markets/sync`

## Способы авторизации

* **API-ключ** через `X-API-Key: <key>` или `Authorization: Bearer <key>`.
* **Белый список сетей (CIDR)** — запрос пускается по IP клиента, если он попадает в список.

Доступ даётся, если выполнено **(ключ ИЛИ CIDR)**. (Можно переключить на «И ключ, и CIDR».)

## Переменные окружения

* `ADMIN_API_KEY` — строка ключа (обязателен, если не используешь CIDR).
* `ADMIN_TRUSTED_CIDRS` — список подсетей через запятую (пример: `127.0.0.1/32,::1/128,10.0.0.0/8`).
* `ADMIN_REQUIRE_BOTH` — если `1|true`, требовать **и** ключ, **и** попадание в CIDR.

> Значения читаются при **старте** сервиса. Меняешь — перезапускай процесс/контейнер.

## Локальный запуск (без Docker)

```bash
export DB_DSN='postgres://postgres:pass@localhost:5432/tickers?sslmode=disable'
export ADMIN_API_KEY='supersecret'
export ADMIN_TRUSTED_CIDRS='127.0.0.1/32,::1/128'
./app
```

Проверка:

```bash
# с ключом (любой IP)
curl -H 'X-API-Key: supersecret' -X POST http://localhost:8080/admin/markets/sync

# или по локальному IP без ключа (если CIDR включает localhost)
curl -X POST http://localhost:8080/admin/markets/sync
```

Если всё ок — получишь JSON с `summary`. Если видишь `{"error":"admin endpoint restricted"}`, значит ключ не совпал **и** IP не в списке.

# API документация

Все ответы и примеры — актуальны для текущей реализации сервиса.

## Общие сведения

* База URL: `http://<host>:8080`
* Ошибки: JSON вида `{"error":"<сообщение>"}`.
* Статусы:

  * `200 OK` — успех.
  * `400 Bad Request` — ошибка запроса (нехватка параметров и т.п.).
  * `403 Forbidden` — доступ запрещён (админ-ручки).
  * `500 Internal Server Error` — внутренняя ошибка.

Справочник обозначений:

* **Биржевые слаги**: `binance`, `bybit`, `okx`, `coinbase`, `upbit`, `bithumb`, `robinhood` (может быть отключена).
* **Слаги списков**: `"<source>_to_<target>"`, напр. `okx_to_upbit`.
* Формат строки списка: `"spot, futures"`; если фьючерса нет — `"spot, none"`.

---

## 1) Health-check

### `GET /health`

Краткая диагностика сервиса и зависимостей.

**Ответ 200 (пример):**

```json
{
  "status": "ok",
  "version": "0.1.0",
  "commit": "local",
  "buildTime": "2025-08-14T12:00:00Z",
  "uptime": "5m53s",
  "checks": { "db": "ok" },
  "now": "2025-08-14T18:01:21Z"
}
```

**Пример:**

```bash
curl -s http://localhost:8080/health | jq .
```

---

## 2) Админ: синхронизация рынков (Snapshot)

> Закрыт по API-ключу/белому списку сетей.

### `POST /admin/markets/sync`

Синхронизирует локальную таблицу `markets` с актуальными данными бирж-источников (для всех активных бирж). Идемпотентно.

**Аутентификация (любой из вариантов):**

* Заголовок `X-API-Key: <ключ>`, или
* `Authorization: Bearer <ключ>`, или
* IP клиента входит в `ADMIN_TRUSTED_CIDRS`.

**ENV:**

* `ADMIN_API_KEY` — ключ.
* `ADMIN_TRUSTED_CIDRS` — CIDR-список (через запятую).
* `ADMIN_REQUIRE_BOTH=1` — требовать и ключ, и IP.

**Ответ 200 (пример):**

```json
{
  "summary": {
    "1": [0, 3757, 2],
    "2": [2, 0, 34],
    "3": [0, 1011, 2],
    "4": [0, 0, 260],
    "5": [0, 0, 0],
    "6": [0, 0, 15],
    "7": [0, 0, 0]
  }
}
```

Значения: `[added, updated, archived]` по `exchange_id`.

**Примеры:**

```bash
# с ключом
curl -s -H 'X-API-Key: supersecret' -X POST http://localhost:8080/admin/markets/sync | jq .

# через Bearer
curl -s -H 'Authorization: Bearer supersecret' -X POST http://localhost:8080/admin/markets/sync | jq .
```

---

## 3) Обновление списков (build + save)

### `POST /update`

Делает полный цикл:

1. синхронизирует рынки (`markets`) для активных бирж,
2. пересобирает и **транзакционно** перезаписывает списки (`list_items`) по выбранным фильтрам.

**Query-параметры (необязательные):**

* `source` — один слаг источника (напр. `okx`). Если не задан — все источники.
* `target` — один слаг цели (напр. `upbit`). Если не задан — все цели.

> Порядок: синк рынков → пересборка соответствующих списков.

**Ответ 200 (пример):**

```json
{
  "markets_sync": {
    "1": [0, 0, 0],
    "2": [0, 0, 0],
    "3": [0, 0, 0],
    "4": [0, 0, 0],
    "5": [0, 0, 0],
    "6": [0, 0, 0],
    "7": [0, 0, 0]
  },
  "lists_updated": {
    "okx_to_upbit": 175
  }
}
```

`lists_updated` — количество записанных строк на каждый список.

**Примеры:**

```bash
# все источники и цели
curl -s -X POST 'http://localhost:8080/update' | jq .

# только OKX → Upbit
curl -s -X POST 'http://localhost:8080/update?source=okx&target=upbit' | jq .

# все источники → Coinbase
curl -s -X POST 'http://localhost:8080/update?target=coinbase' | jq .
```

---

## 4) Публичные списки

### `GET /api/lists/:slug`

Возвращает сформированный список по слагу (`<source>_to_<target>`).

**Query-параметры:**

* `as_text` — если `1|true|yes`, ответ в `text/plain` построчно.

**Ответ 200 (JSON):**

```json
{
  "slug": "okx_to_upbit",
  "items": [
    "ACA-USDT, none",
    "ACE-USDT, none",
    "ACH-USDT, none"
  ]
}
```

**Ответ 200 (TEXT при `as_text=1`):**

```
ACA-USDT, none
ACE-USDT, none
ACH-USDT, none
```

**Примеры:**

```bash
curl -s 'http://localhost:8080/api/lists/okx_to_upbit' | jq .
curl -s 'http://localhost:8080/api/lists/okx_to_upbit?as_text=1' | head -n 20
```

---

### `GET /api/lists?target=<slug>`

Возвращает **все списки** для указанной целевой биржи, сгруппированные по источникам.

**Query-параметры (обязательные):**

* `target` — слаг целевой биржи, напр. `upbit`.

**Доп. параметр:**

* `as_text` — если `1|true|yes`, ответ в `text/plain` (плоский построчный список, источники отсортированы по алфавиту).

**Ответ 200 (JSON):**

```json
{
  "target": "upbit",
  "sources": {
    "okx": [
      "ACA-USDT, none",
      "ACE-USDT, none"
    ],
    "binance": [
      "AAAUSDT, AAAUSDT-PERP"
    ]
  }
}
```

**Ответ 200 (TEXT при `as_text=1`):**

```
AAAUSDT, AAAUSDT-PERP
ACA-USDT, none
ACE-USDT, none
...
```

**Примеры:**

```bash
curl -s 'http://localhost:8080/api/lists?target=upbit' | jq .
curl -s 'http://localhost:8080/api/lists?target=upbit&as_text=1' | head -n 30
```

---

## Замечания по поведению

* **Идемпотентность**: повторный вызов `/admin/markets/sync` или `/update` может возвращать нули (данные не изменились).
* **Отключённые биржи**: списки для пар с биржами, у которых `exchanges.is_active=false`, не формируются. Флаг настраивается в БД (миграция добавлена).
* **Формат строк**: `"spot, futures"`; если фьючерса нет — `"spot, none"`. Строки отсортированы по `spot_ticker`.
* **Фильтры**:

  * `source`/`target` в `/update` — одиночные значения (если передать CSV, берётся первый элемент).
  * Для Upbit/Bithumb действует правило **BTC-only не считается присутствием** (в логике фильтрации это учтено).
  * Для Binance действует правило: если на цели у монеты есть **и** спот, **и** фьючерс — не исключаем из источника.

---

## Быстрые команды для проверки

```bash
# Health
curl -s http://localhost:8080/health | jq .

# Синхронизация рынков (админ)
curl -s -H 'Authorization: Bearer supersecret' -X POST http://localhost:8080/admin/markets/sync | jq .

# Обновление всех списков
curl -s -X POST 'http://localhost:8080/update' | jq .

# Обновление конкретного списка
curl -s -X POST 'http://localhost:8080/update?source=okx&target=upbit' | jq .

# Публичный список по слагу (TEXT)
curl -s 'http://localhost:8080/api/lists/okx_to_upbit?as_text=1' | head -n 20

# Публичные списки по цели (JSON)
curl -s 'http://localhost:8080/api/lists?target=upbit' | jq .
```

Этого достаточно, чтобы интегрировать сервис и адекватно мониторить его состояние.
