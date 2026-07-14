# Prompt: Backend для Telegram Mini App (LifeOS)

**Назначение:** отдельный этап для бэкенд-агента / сессии.  
**Фронтенд в этом этапе не трогать** — только Go API, деплой, туннель, BotFather WebApp URL.  
**Связано:** [UX_UI_PLAN.md](UX_UI_PLAN.md) · Phase A (`MA-A1`, `MA-A4`)

---

## Контекст инфраструктуры

- LifeOS крутится **локально** (Mac / домашняя машина): `lifeos serve` + Postgres.
- Есть **статический публичный IP**.
- Можно: **пробросить порты** на роутере **или** поднять туннель (**ngrok**, Cloudflare Tunnel, localtunnel, frp).
- Telegram Mini App требует **HTTPS** URL для WebApp (Http только для очень узких исключений — не рассчитывать).

---

## Цель этапа

Сделать так, чтобы Mini App из Telegram ходил на твой backend:

1. Пользователь открывает WebApp из бота → HTTPS.
2. `initData` валидируется на сервере → JWT.
3. REST `/api/v1/*` доступен с того же origin (или CORS настроен).
4. Статика Mini App (`web/miniapp/dist`) отдаётся с `/app/`.

---

## Варианты доступа (выбрать один)

### A. Статический IP + порт + reverse proxy (предпочтительно для «навсегда»)

```
Internet (443 HTTPS)
    → Router DNAT :443 → home-server:443
        → Caddy / nginx / Traefik
            → /api/*     → localhost:8080
            → /app/*     → static dist (или Vite preview)
            → /webhook/* → localhost:8080  (если webhook mode)
```

**Нужно:**
- Домен (или IP + TLS через Cloudflare / Let's Encrypt DNS challenge).
- Открыть **только 443** наружу (не светить Postgres / 8080 напрямую).
- TLS сертификат обязателен.

### B. ngrok / Cloudflare Tunnel (быстрый dogfood)

```bash
# пример ngrok
ngrok http 8080
# или туннель на Caddy, который уже проксирует :8080 + /app
```

**Плюсы:** HTTPS сразу, без роутера.  
**Минусы:** URL меняется (free tier), лимиты; для BotFather нужно обновлять Web App URL.

### C. Гибрид

- API на static IP:443  
- Или туннель только на время разработки фронта.

---

## Чеклист бэкенд-работ (порядок)

### 1. Auth для WebApp — `MA-A1` (P0)

Реализовать:

```
POST /api/v1/auth/telegram-webapp
Body: { "init_data": "<WebApp.initData>" }
→ 200 { "access_token", "expires_in", "token_type": "Bearer" }
```

Логика:
- Проверить HMAC подпись `init_data` по Bot Token (Telegram WebApp docs).
- Вытащить `user.id` → EnsureUser / GetUser.
- Выдать тот же JWT, что и `/auth/token`.

Без этого фронт в проде не зайдёт (сейчас клиент уже ждёт этот путь).

### 2. Finance overview — `MA-A4` (P0)

```
GET /api/v1/finance/overview?period=YYYY-MM
→ income_cents, expense_cents, net_cents, currency, categories[]
```

Категории: name, amount_cents, percent, color_hint (optional).  
Фронт уже умеет fallback на cash-flow, но legend пустой без overview.

### 3. Отдача Mini App статики

Вариант:
- Собрать `cd web/miniapp && npm run build` → `dist/`
- Раздавать под `base: /app/` (уже в `vite.config.ts`)
- Либо отдельный static host / Caddy `file_server`

### 4. Reverse proxy + TLS

Минимальный Caddyfile-скелет:

```
lifeos.example.com {
  handle /api/* {
    reverse_proxy localhost:8080
  }
  handle /webhook/* {
    reverse_proxy localhost:8080
  }
  handle /app/* {
    root * /path/to/web/miniapp/dist
    try_files {path} /app/index.html
    file_server
  }
  # опционально health
  handle /health {
    reverse_proxy localhost:8080
  }
}
```

Для ngrok: туннелить уже прокси на `:443`/`caddy`, не сырой 8080 без `/app`.

### 5. BotFather / бот

- Создать/обновить Web App URL: `https://<host>/app/`
- Кнопка Menu Button / inline `web_app` с этим URL.
- Если polling: webhook не обязателен.
- Если позже webhook: `LIFEOS_TELEGRAM_MODE=webhook` + публичный HTTPS.

### 6. Env на сервере

```
TELEGRAM_BOT_TOKEN=...
LIFEOS_JWT_SECRET=...
LIFEOS_API_KEY=...          # только для dev token auth
LIFEOS_DATABASE_URL=...
# webhook только если нужен:
# LIFEOS_TELEGRAM_MODE=webhook
# LIFEOS_TELEGRAM_WEBHOOK_URL=https://host/webhook/telegram
# LIFEOS_TELEGRAM_WEBHOOK_SECRET=...
```

Postgres **не** публиковать в интернет.

### 7. CORS (если фронт и API на разных origin)

Предпочтительно same-origin через proxy.  
Если раздельно — разрешить origin Mini App + credentials/Authorization header.

### 8. Проверка

1. `curl https://host/health`  
2. Открыть Mini App из Telegram на телефоне.  
3. Auth проходит (не «не удалось войти»).  
4. Tasks today + finance грузятся.  
5. HTTPS валиден (без предупреждений).

---

## Безопасность (обязательно)

| Делать | Не делать |
|--------|-----------|
| TLS на входе | Открывать `:5432` наружу |
| Только 443 (или туннель) | Светить `:8080` без proxy |
| Валидировать initData HMAC | Доверять `telegram_id` из body без подписи |
| Короткий JWT TTL + refresh/re-auth через initData | Вечный token в localStorage без плана |
| Firewall: allow 443 only | API key в клиентском бандле для прода |

---

## Acceptance критерий этапа Backend

- [ ] `POST /api/v1/auth/telegram-webapp` работает с реальным initData  
- [ ] `GET /api/v1/finance/overview` отдаёт категории за период  
- [ ] `https://<host>/app/` открывается в Telegram WebApp  
- [ ] API same-origin или CORS ок  
- [ ] Postgres и JWT secret не торчат наружу  

---

## Что НЕ входит в этот этап

- Редизайн UI, новые React-экраны, Vue/Next миграции — это фронтенд.  
- Полный feature parity доменов — Phase B/C в UX_UI_PLAN.

## Промпт для бэкенд-агента (короткий)

```
Ты бэкенд-инженер LifeOS (Go). Реализуй этап Backend Mini App по docs/miniapp/BACKEND_PROMPT.md.

Инфра: локальный сервер, статический IP; можно port-forward или ngrok/Cloudflare Tunnel.
Обязательно: HTTPS для Telegram WebApp.

Сделать по порядку:
1) POST /api/v1/auth/telegram-webapp (HMAC initData → JWT)
2) GET /api/v1/finance/overview?period=
3) Раздача web/miniapp/dist на /app/ + reverse proxy
4) Инструкция BotFather URL и env
5) Не трогать React UI кроме .env.example если нужно

Критерии приёмки — в BACKEND_PROMPT.md § Acceptance.
```
