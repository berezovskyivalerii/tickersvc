# CI/CD и бэкапы

Этот проект разворачивается через GitLab CI/CD. Основные этапы: тесты → сборка Docker-образа → миграции → деплой. Бэкапы БД — отдельным джобом/расписанием.

## Структура пайплайна

1. **test** — `go test ./... -race`, публикует покрытие.
2. **docker-build** — собирает и пушит образ в GitLab Container Registry.
   - Теги:
     - `:<commit-sha>` всегда,
     - `:latest` для `main`,
     - `:<git-tag>` для релизов.
3. **migrate** — применяет SQL-миграции (`goose up`) к базе `DB_DSN`.
4. **deploy** — заходит по SSH на сервер, делает `docker compose pull && up -d`.
5. **db_backup** — создаёт `pg_dump` (`.dump`) и сохраняет как artifact на 14 дней.

## Переменные окружения (GitLab → Settings → CI/CD → Variables)

Обязательные:
- `DB_DSN` — `postgres://user:pass@host:5432/tickers?sslmode=disable`
- `SSH_HOST`, `SSH_USER`, `SSH_PRIVATE_KEY` — для деплоя
- `REMOTE_APP_DIR` — каталог на сервере с `docker-compose.yml`

Опциональные:
- `PUBLIC_HOST` — адрес сервиса (отображается в среде production)
- `RUN_DB_BACKUP` — установи `1`, чтобы вручную запустить `db_backup` из пайплайна (без расписания)

Registry: GitLab предоставляет `CI_REGISTRY_*` автоматически.

## Сервер деплоя

На сервере должен быть установлен Docker и `docker compose`. В каталоге `$REMOTE_APP_DIR` лежит файл `docker-compose.yml`, который использует образ из `CI_REGISTRY_IMAGE` (например, `image: registry.gitlab.com/namespace/project:latest`).

Команда, которую запускает деплой:
```sh
docker login $CI_REGISTRY -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD
cd $REMOTE_APP_DIR
docker compose pull
docker compose up -d
docker image prune -f
