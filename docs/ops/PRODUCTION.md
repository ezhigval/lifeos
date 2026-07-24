# Production hosting — LifeOS Mini App + bot (без trycloudflare)

Quick-tunnel (`*.trycloudflare.com`) удобен на час, но URL умирает → бот/Mini App «молчат».
Ниже стабильный HTTPS + деплой через **GitHub Actions**.

> **Важно:** GitHub Actions — это CI/CD (сборка и выкладка), а не место, где бот живёт 24/7.
> Бот крутится на **Fly.io** (или VPS). Actions только деплоит туда.

## Что выбрать

| Вариант | Когда | URL |
|--------|--------|-----|
| **A. Fly.io + GitHub Actions** | Быстрый прод без своего сервера | `https://<app>.fly.dev/app/` |
| **B. VPS + Caddy** | Есть сервер + домен (`lifeos.…`) | `https://lifeos.example.com/app/` |

Оба отдают **один origin**: `/app` + `/api` + `/webhook` — Telegram WebView это любит.

---

## A. Fly.io через GitHub Actions (рекомендуется)

### Один раз (локально)

```bash
# 1) Аккаунт: https://fly.io  →  fly auth login
fly apps create lifeos
fly postgres create --name lifeos-db --region ams
fly postgres attach lifeos-db -a lifeos
fly tokens create deploy -x 999999h   # → это FLY_API_TOKEN
```

### Secrets в GitHub

Repo → **Settings → Secrets and variables → Actions** → New repository secret:

| Secret | Откуда |
|--------|--------|
| `FLY_API_TOKEN` | `fly tokens create deploy` |
| `TELEGRAM_BOT_TOKEN` | @BotFather |
| `LIFEOS_JWT_SECRET` | длинная случайная строка |
| `LIFEOS_API_KEY` | случайная строка |
| `LIFEOS_TELEGRAM_WEBHOOK_SECRET` | `openssl rand -hex 24` |
| `LIFEOS_SEED_TELEGRAM_ID` | опционально, твой telegram id |
| `FLY_APP` | опционально, default `lifeos` |

### Деплой

- **Actions → Deploy → Run workflow**, или
- push в `main` (workflow `.github/workflows/deploy.yml`)

Скрипт: `scripts/ci-deploy-fly.sh` — secrets на Fly, `fly deploy`, health, `setWebhook` + Mini App menu.

Проверка:

```bash
curl -sS https://lifeos.fly.dev/health
curl -sS https://lifeos.fly.dev/app/ | head
```

В Telegram: **Menu → Mini App** или свежая синяя кнопка (не из старой истории).

Локальный деплой без Actions: `./scripts/deploy-fly.sh` (нужен `.env` + flyctl).

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
