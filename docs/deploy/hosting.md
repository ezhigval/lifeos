# Другие способы публичного HTTPS

На своей ВМ, где входящий 80/443 нестабилен, используй [vm.md](vm.md) и [EDGE.md](EDGE.md). Там вход — Cloudflare, не nginx и не Caddy. Fly.io не используется. Ниже вариант, где приложение само слушает публичный адрес.

## VPS и Caddy

Нужны DNS A-запись на машину и открытые 80/443.

```bash
cp deployments/.env.prod.example deployments/.env.prod
# LIFEOS_DOMAIN, CADDY_ACME_EMAIL, TELEGRAM_BOT_TOKEN,
# LIFEOS_JWT_SECRET (от 32 байт), POSTGRES_PASSWORD
./scripts/deploy-compose.sh
```

Файлы: `deployments/docker-compose.prod.yml`, `deployments/Caddyfile`. Webhook URL compose собирает из `LIFEOS_DOMAIN`.
