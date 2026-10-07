# Другие способы публичного HTTPS

На своей ВМ, где входящий 80/443 нестабилен, используй [vm.md](vm.md). Ниже два варианта, где приложение само слушает публичный адрес.

## Fly.io

GitHub Actions выкладывает на Fly при push в `main` (workflow `.github/workflows/deploy.yml`). Секреты репозитория: `FLY_API_TOKEN`, `TELEGRAM_BOT_TOKEN`, `LIFEOS_JWT_SECRET`, `LIFEOS_API_KEY`, `LIFEOS_TELEGRAM_WEBHOOK_SECRET`. Опционально `FLY_APP` и `LIFEOS_SEED_TELEGRAM_ID`.

Локально, без Actions:

```bash
fly postgres create --name lifeos-db --region ams
fly postgres attach lifeos-db -a lifeos
./scripts/deploy-fly.sh
```

Скрипт собирает образ из `deployments/Dockerfile`, кладёт секреты из локального `.env` и вызывает `scripts/set-telegram-urls.sh` (кнопка меню и webhook). `fly postgres attach` пишет `DATABASE_URL`. Приложение читает его, если `LIFEOS_DATABASE_URL` пуст.

```bash
curl -fsS https://<app>.fly.dev/health
```

## VPS и Caddy

Нужны DNS A-запись на машину и открытые 80/443.

```bash
cp deployments/.env.prod.example deployments/.env.prod
# LIFEOS_DOMAIN, CADDY_ACME_EMAIL, TELEGRAM_BOT_TOKEN,
# LIFEOS_JWT_SECRET (от 32 байт), POSTGRES_PASSWORD
./scripts/deploy-compose.sh
```

Файлы: `deployments/docker-compose.prod.yml`, `deployments/Caddyfile`. Webhook URL compose собирает из `LIFEOS_DOMAIN`.
