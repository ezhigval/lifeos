# Другие способы публичного HTTPS

На своей ВМ, где входящий 80/443 нестабилен, используй [vm.md](vm.md) и [EDGE.md](EDGE.md). Там вход — Cloudflare. Fly.io не используется. Ниже вариант, где приложение само слушает публичный адрес.

## VPS и Caddy

Нужны DNS A-запись на машину и открытые 80/443.

Первая загрузка VDS Timeweb (Ubuntu 24.04): вставить `deployments/timeweb/user-data.sh` в User data. Скрипт ставит Docker, открывает 22/80/443 и пишет секреты в `/opt/lifeos/.env`. Токен бота остаётся пустым, контейнеры не стартуют.

```bash
cp deployments/.env.prod.example deployments/.env.prod
# LIFEOS_DOMAIN, CADDY_ACME_EMAIL, TELEGRAM_BOT_TOKEN,
# LIFEOS_JWT_SECRET (от 32 байт), POSTGRES_PASSWORD
./scripts/deploy-compose.sh
```

Файлы: `deployments/docker-compose.prod.yml`, `deployments/Caddyfile`. Webhook URL compose собирает из `LIFEOS_DOMAIN`.
