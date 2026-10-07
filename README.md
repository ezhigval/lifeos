# LifeOS

Personal operating system for tasks, projects, money, habits, and health. The Telegram bot is the main interface. The same process serves a REST API and a Telegram Mini App. Domain code does not know which client called it.

Персональная операционная система: задачи, проекты, деньги, привычки, здоровье. Основной интерфейс — Telegram. Тот же процесс отдаёт REST и Mini App.

**Альфа 1.0** (`LifeOS_alpha_1.0.0`). Лицензия [MIT](LICENSE).

## Что умеет

- Задачи, проекты, сферы жизни, план дня, обзоры
- Финансы: доходы, расходы, долги, план
- Привычки, календарь, заметки, здоровье, карьера, напоминания
- Бот понимает команды и свободный текст. LLM подключается отдельно и не обязателен
- Mini App на том же API

Как это разложено по модулям: [docs/architecture/ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md).

## Поставить себе

Нужны Docker, токен бота от [@BotFather](https://t.me/BotFather) и секрет JWT длиннее 32 байт.

```bash
git clone https://github.com/ezhigval/lifeos.git
cd lifeos
cp .env.example .env
# TELEGRAM_BOT_TOKEN и LIFEOS_JWT_SECRET

make docker-up
curl -fsS http://127.0.0.1:8080/health
```

Приложение слушает `:8080`. Mini App: `http://127.0.0.1:8080/app/`. Postgres с хоста: порт `5433`. Образ сам накатывает миграции.

Напиши боту `/start`. Без публичного HTTPS откроется только чат. Mini App из Telegram требует HTTPS: для ноутбука это [короткий туннель](docs/telegram.md), для сервера — [домен и named tunnel](docs/deploy/vm.md).

Дальше по шагам: [docs/getting-started.md](docs/getting-started.md).

## Десктоп-пакет

```bash
make package          # эта машина
make package-mac      # darwin/arm64
make package-linux    # linux/amd64
make package-win      # windows/amd64
```

Архив `dist/LifeOS_alpha_1.0.0_<os>_<arch>.tar.gz` содержит бинарь, Mini App, миграции и скрипты Start / Stop / Logs / Settings. Postgres ставится отдельно. Подробности внутри пакета, в `README.txt`.

## Документация

| | |
|---|---|
| Запуск локально и в Docker | [docs/getting-started.md](docs/getting-started.md) |
| Бот, Mini App, туннель на ноутбуке | [docs/telegram.md](docs/telegram.md) |
| ВМ, секреты, автодеплой, постоянный туннель | [docs/deploy/vm.md](docs/deploy/vm.md) |
| VPS и Caddy | [docs/deploy/hosting.md](docs/deploy/hosting.md) |
| Архитектура | [docs/architecture/ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md) |
| Модель и схема | [docs/architecture/DOMAIN_MODEL.md](docs/architecture/DOMAIN_MODEL.md), [docs/diagrams/](docs/diagrams/) |
| Решения | [docs/adr/](docs/adr/) |
| Фразы бота и агент | [docs/ai/SCENARIOS.md](docs/ai/SCENARIOS.md), [docs/ai/AGENT.md](docs/ai/AGENT.md) |
| LLM | [docs/ops/LLM.md](docs/ops/LLM.md) |
| OpenAPI | [docs/api/openapi.yaml](docs/api/openapi.yaml) |
| Как контрибьютить | [CONTRIBUTING.md](CONTRIBUTING.md) |

Оглавление: [docs/README.md](docs/README.md).

## Стек

Go 1.25, Chi, PostgreSQL 16, pgx, sqlc, goose, Cobra. Mini App: React 19, Vite, Tailwind CSS 4. По желанию: Ollama или OpenAI-совместимый LLM, Prometheus, OpenTelemetry.

## Разработка

```bash
make dev          # сервер на машине, Postgres уже должен быть
make test
make lint         # golangci-lint
make ci
```

Секреты в репозиторий не коммитятся. Шаблоны: `.env.example`. Проверка: `bash scripts/check-no-secrets.sh`.

## Лицензия

[MIT](LICENSE).
