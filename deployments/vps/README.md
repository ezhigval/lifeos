# ВPS-деплой (Yandex Cloud Compute и т.п.)

Один скрипт поднимает всё с нуля на чистой Ubuntu 22.04/24.04 от root:
PostgreSQL + роль/БД `lifeos`, Go 1.25, cloudflared, nginx (прокси :80 → :8080),
UFW (только SSH+HTTP), два systemd-сервиса (`lifeos`, `lifeos-tunnel`), миграции при старте.

## Быстрый старт (ключ-пароль)
```bash
ssh-copy-id root@<IP>            # или sshpass -p '<pass>' ssh-copy-id ...
scp deployments/vps/install.sh root@<IP>:/root/
ssh root@<IP>
REPO_URL=<git-адрес репо> bash /root/install.sh
nano /opt/lifeos/lifeos.env      # LIFEOS_TELEGRAM_BOT_TOKEN + LIFEOS_MINIAPP_URL=<туннель>/app/
systemctl restart lifeos
journalctl -u lifeos-tunnel -f   # скопировать trycloudflare-URL
curl https://<туннель>/health
```

## Обновления с dev-бокса (без git-ключа на сервере)
```bash
rsync -az --delete -e "ssh $SSH_OPTS" /workspace/ root@<IP>:/opt/lifeos/src/
ssh root@<IP> 'cd /opt/lifeos/src && /usr/local/go/bin/go build -o /opt/lifeos/bin/lifeos ./cmd/lifeos && systemctl restart lifeos'
```
(локально перед rsync: `cd web/miniapp && npm run build`)

## Примечания
- Quick-туннель trycloudflare меняет хост при рестарте сервиса → обновить `LIFEOS_MINIAPP_URL` и отправить боту `/start`. Для постоянного домена — named tunnel (`cloudflared tunnel login`).
- `PASTE-BOT-TOKEN`/`PASTE-TUNNEL-HOST` в создаваемом `.env` — единственные поля, которые нужно вписать руками; секреты JWT/API генерируются автоматически.
- Проверено локально: bash -n ОК. Реальный прогон — на машине (ждём IP/ключ).

## Telegram-эгресс через Cloudflare Worker (tg-proxy) — бесплатно, внутри ВМ

Yandex Cloud блокирует исходящие к api.telegram.org, но **не** блокирует Cloudflare.
Схема: приложение → локальный forward-proxy `127.0.0.1:8081` (tg-proxy.py) → ваш Cloudflare
Worker (`*.workers.dev`) → api.telegram.org. Всё бесплатно (тариф Workers 100k запросов/день).

1. Задеплойте воркер (один раз, с любой машины): dash.cloudflare.com → Workers → Create →
   вставьте код `deployments/vps/tg-proxy-worker.js` → Deploy. Или:
   `cd deployments/vps && npx --yes wrangler deploy` (потребуется `npx wrangler login`).
   Запомните URL: `https://tg-proxy.<ваш-сабдомен>.workers.dev`.
2. На ВМ:
   ```bash
   scp deployments/vps/tg-proxy.py deployments/vps/tg-proxy.sh <user>@<VM>:/tmp/
   ssh <user>@<VM>
   sudo LIFEOS_TG_PROXY_WORKER_URL=https://tg-proxy.<sub>.workers.dev bash /tmp/tg-proxy.sh install
   sudo bash /tmp/tg-proxy.sh test   # getMe через прокси
   ```
3. Приложение само подхватит `LIFEOS_HTTP_PROXY` из `/opt/lifeos/lifeos.env` (сервис перезапускается).
   После этого работают polling и webhook (webhook тоже проходит: TG→Cloudflare→туннель — входящее).
4. Отключить: `sudo bash /tmp/tg-proxy.sh remove`.

Примечание: base URL клиента Telegram намеренно `http://` — plain-forward proxy ретранслирует
absolute-form URI без CONNECT/TLS-туннелирования; сам хоп до Worker идёт по HTTPS внутри tg-proxy.py.
