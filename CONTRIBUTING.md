# Участие в разработке

LifeOS — модульный монолит на Go. Бизнес-правила живут в `internal/<context>/domain` и `app`. HTTP и Telegram только вызывают эти сценарии.

## Окружение

- Go 1.25
- Node.js 22 (Mini App)
- PostgreSQL 16 или Docker
- [golangci-lint](https://golangci-lint.run) v2, если гоняешь `make lint`

```bash
cp .env.example .env
make docker-up
make test
```

`make docker-up` поднимает Postgres и приложение. Образ сам применяет миграции перед `serve`.

## Куда класть код

| Слой | Каталог | Правило |
|------|---------|---------|
| Домен | `internal/<context>/domain` | Без импорта `app`, `infra`, `transport` |
| Сценарии | `internal/<context>/app` | Вызывает домен, не знает про HTTP |
| Хранилище | `internal/<context>/infra` | PostgreSQL, sqlc |
| Транспорт | `internal/transport` | Chi и Telegram |

Новая таблица — файл в `migrations/` с маркерами `-- +goose Up` и `-- +goose Down`. Контракт REST — `docs/api/openapi.yaml`. Проверка: `make openapi-check`.

Решения, которые меняют границы модулей, записываются в `docs/adr/`.

## Проверки перед PR

```bash
make ci
cd web/miniapp && npm ci && npm run build
bash scripts/check-no-secrets.sh
```

`make ci` включает tidy, lint, OpenAPI, тесты, порог покрытия и сборку. Для lint нужен golangci-lint.

## Секреты

В git не попадают `.env`, токен бота, JWT, ключи SSH, `TUNNEL_TOKEN`, `PROXY_SECRET`. На ВМ они лежат в `/opt/lifeos/.env` и `/opt/lifeos/secrets` с правами `600`. Пустые примеры — `.env.example` и `deployments/cloudflared/tunnel.env.example`.

## Стиль

- Go: `gofmt`. Комментарий только там, где из кода не ясно зачем.
- Пользовательские строки бота и Mini App — по-русски.
- Не добавляй второй способ сделать то же самое, если скрипт или документ уже есть. Сначала ссылка, потом новый файл.
