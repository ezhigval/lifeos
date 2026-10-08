# Запуск

Нужны Go 1.25, Node.js 22 и PostgreSQL 16. Для обычного старта хватает Docker.

## Docker

```bash
git clone https://github.com/ezhigval/lifeos.git
cd lifeos
cp .env.example .env
```

В `.env` заполни:

- `TELEGRAM_BOT_TOKEN` — токен от [@BotFather](https://t.me/BotFather)
- `LIFEOS_JWT_SECRET` — не короче 32 байт (`openssl rand -hex 32`)

```bash
make docker-up
curl -fsS http://127.0.0.1:8080/health
```

Compose публикует Postgres на `127.0.0.1:5433` и приложение на `127.0.0.1:8080`. Образ перед `serve` выполняет `migrate up`. Mini App отдаётся с `http://127.0.0.1:8080/app/`.

Остановить: `make docker-down`.

Метрики, Grafana и Jaeger:

```bash
make observability-up
```

## Без Docker

Postgres должен слушать адрес из `LIFEOS_DATABASE_URL` (в примере это `localhost:5433`).

```bash
cp .env.example .env
make migrate-up
make dev
```

Сборка Mini App в статику, которую раздаёт сервер:

```bash
make miniapp-build
```

Каталог задаётся `LIFEOS_STATIC_DIR` (в примере `web/miniapp/dist`).

Фронт отдельно, с горячей перезагрузкой:

```bash
make miniapp-dev
```

## Десктоп-пакет alpha 1.0

Каталог с бинарём, Mini App, миграциями и кнопками запуска:

```bash
make package          # ОС и архитектура этой машины
make package-mac      # darwin/arm64
make package-linux    # linux/amd64
make package-win      # windows/amd64
```

Результат: `dist/LifeOS_alpha_1.0.0_<os>_<arch>/` и рядом `.tar.gz` / `.zip`. Внутри `Start`, `Stop`, `Logs`, `Settings`. Postgres пакет не включает. Как им пользоваться, написано в `README.txt` внутри архива.

## Проверка

```bash
make test
make openapi-check
```

Полный прогон, как в CI: `make ci` (нужен golangci-lint).

Дальше: [бот и Mini App](telegram.md). Если нужен адрес, который не меняется, — [ВМ и туннель](deploy/vm.md).
