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

## MVP scope (фронт)

| Экран | Статус |
|-------|--------|
| Home: задачи + finance | ✅ |
| Сферы → проекты → задачи / task detail | ✅ |
| Привычки / Календарь / Напоминания | ✅ |
| Аналитика / Заметки / Здоровье / Карьера / Долги | ✅ |
| Настройки: обзоры, quiet hours, CRUD сфер | ✅ |

Lead prompt: [FRONTEND_LEAD_PROMPT.md](FRONTEND_LEAD_PROMPT.md)

## Нужно с бэкенда для продакшена Telegram

См. [BACKEND_PROMPT.md](BACKEND_PROMPT.md):

1. `POST /api/v1/auth/telegram-webapp` ✅
2. `GET /api/v1/finance/overview?period=` ✅
3. HTTPS раздача `/app/` + proxy `/api`

## Tunnel / Cloudflare Error 1033

**1033** = public `*.trycloudflare.com` hostname exists, but **origin `:8080` is down** or the Mini App button still points at an **old** tunnel URL.

```bash
# 1) App must answer locally
curl -sS http://127.0.0.1:8080/health

# 2) Full stack + fresh tunnel URL + recreate app
make stack-up

# 3) In Telegram: /start  (forces new reply keyboard with current LIFEOS_MINIAPP_URL)

# 4) Verify initData → telegram_id → JWT
make verify-webapp-auth
```

Do **not** open an old Mini App link from chat history after `cloudflared` restarted — the hostname dies → 1033.

## Auth path (initData → telegram user id)

1. Client freezes `Telegram.WebApp.initData` (or `#tgWebAppData`) before Router
2. `POST /api/v1/auth/telegram-webapp` with `{ "init_data": "..." }`
3. Server HMAC-validates (`secret = HMAC_SHA256(bot_token, key=WebAppData)`)
4. Parses signed `user` JSON → `id` (telegram_id)
5. `EnsureUserByTelegram(telegram_id)` → LifeOS UUID
6. JWT `sub` = LifeOS user_id; response also echoes `telegram_id` for session bind
