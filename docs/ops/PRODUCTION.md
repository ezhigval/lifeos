# Production hosting — LifeOS Mini App + bot (без trycloudflare)

Quick-tunnel (`*.trycloudflare.com`) удобен на час, но URL умирает → бот/Mini App «молчат».
Ниже два нормальных варианта со **стабильным HTTPS**.

## Что выбрать

| Вариант | Когда | URL |
|--------|--------|-----|
| **A. Fly.io** | Нет VPS/домена, хочешь быстро | `https://<app>.fly.dev/app/` |
| **B. VPS + Caddy** | Есть сервер + домен | `https://lifeos.example.com/app/` |

Оба отдают **один origin**: `/app` + `/api` + `/webhook` — Telegram WebView это любит.

---

## A. Fly.io

```bash
# 1) Аккаунт: https://fly.io  →  fly auth login
# 2) Postgres (один раз)
fly postgres create --name lifeos-db --region ams
fly postgres attach lifeos-db -a lifeos
# 3) Деплой + привязка URL к Telegram
./scripts/deploy-fly.sh
```

Скрипт:
- собирает Docker-образ из `deployments/Dockerfile`
- выставляет secrets из корневого `.env`
- ставит `LIFEOS_MINIAPP_URL` / webhook
- вызывает `scripts/set-telegram-urls.sh` (menu button + webhook)

`fly postgres attach` пишет `DATABASE_URL` — приложение читает его как fallback к `LIFEOS_DATABASE_URL`.

Проверка:

```bash
curl -sS https://<app>.fly.dev/health
curl -sS https://<app>.fly.dev/app/ | head
```

В Telegram: **Menu → Mini App** или синяя inline-кнопка (не старая из истории).

---

## B. VPS + Docker Compose + Caddy (Let's Encrypt)

```bash
cp deployments/.env.prod.example deployments/.env.prod
# LIFEOS_DOMAIN=lifeos.example.com
# CADDY_ACME_EMAIL=you@example.com
# TELEGRAM_BOT_TOKEN=...
# LIFEOS_JWT_SECRET=...
# LIFEOS_TELEGRAM_WEBHOOK_SECRET=...
# POSTGRES_PASSWORD=...
# LIFEOS_SEED_TELEGRAM_ID=1034074077

# DNS: A-запись LIFEOS_DOMAIN → IP VPS
# Firewall: 80/tcp, 443/tcp

./scripts/deploy-compose.sh
```

Или вручную:

```bash
docker compose -f deployments/docker-compose.prod.yml \
  --env-file deployments/.env.prod up -d --build

export TELEGRAM_BOT_TOKEN=...
export LIFEOS_MINIAPP_URL=https://lifeos.example.com/app/
export LIFEOS_TELEGRAM_WEBHOOK_URL=https://lifeos.example.com/webhook/telegram
export LIFEOS_TELEGRAM_WEBHOOK_SECRET=...
export LIFEOS_NOTIFY_CHAT_ID=1034074077
./scripts/set-telegram-urls.sh
```

Файлы:
- `deployments/docker-compose.prod.yml`
- `deployments/Caddyfile`
- `deployments/.env.prod.example`

---

## Локальная разработка без Mini App

Бот может жить на **polling** без публичного HTTPS:

```bash
# .env
LIFEOS_TELEGRAM_MODE=polling
# LIFEOS_MINIAPP_URL=   # пусто
```

Mini App в этом режиме в Telegram **не откроется** (нужен HTTPS). Для UI крути `npm run dev` на Mac.

---

## Почему «бот жив, Mini App нет»

Частые причины:

1. **Старая кнопка в истории чата** → мёртвый tunnel URL. Лечится свежей inline-кнопкой / Menu после смены URL.
2. **Webhook URL устарел**, а polling не включён → апдейты не доходят (бот «молчит»).
3. Origin `:8080` упал → Cloudflare 530/1033.

`scripts/set-telegram-urls.sh` всегда синхронизирует menu + webhook под актуальный `LIFEOS_MINIAPP_URL`.
