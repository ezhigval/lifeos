# Mini App — локальная разработка

Фронт крутится в браузере без Telegram. Бэкенд-прод (HTTPS / ngrok) — отдельный этап, см. [BACKEND_PROMPT.md](BACKEND_PROMPT.md).

## Быстрый старт

```bash
# 1) Postgres + .env (корень репо)
cp .env.example .env
# LIFEOS_DATABASE_URL=postgres://lifeos:lifeos@localhost:5432/lifeos?sslmode=disable
# LIFEOS_API_KEY=dev-local-api-key
# LIFEOS_JWT_SECRET=dev-jwt-secret-for-local
# LIFEOS_SEED_TELEGRAM_ID=10001

make migrate-up   # или: set -a && source .env && go run ./cmd/lifeos migrate up
set -a && source .env && set +a && go run ./cmd/lifeos serve

# 2) Mini App
cd web/miniapp
cp .env.example .env.local
# VITE_DEV_API_KEY=dev-local-api-key
# VITE_DEV_TELEGRAM_ID=10001
npm install
npm run dev
```

Открыть: **http://localhost:5173/app/**

Vite проксирует `/api` → `:8080`.

## MVP scope (фронт готов)

| Экран | Статус |
|-------|--------|
| Home: задачи + finance capture | ✅ |
| Сферы → проекты → задачи (+ приоритет) | ✅ |
| Привычки (today + track + create) | ✅ |
| Календарь (today + create) | ✅ |
| Ещё / Settings stubs | ✅ |
| TG BackButton / Sheet / errors | ✅ |

## Нужно с бэкенда для продакшена Telegram

См. [BACKEND_PROMPT.md](BACKEND_PROMPT.md):

1. `POST /api/v1/auth/telegram-webapp`
2. `GET /api/v1/finance/overview?period=`
3. HTTPS раздача `/app/` + proxy `/api`
