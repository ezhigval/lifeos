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
- Quick-туннель из этого скрипта меняет хост при рестарте и с текущей Yandex VM не создаётся (`api.trycloudflare.com:443` закрыт). Постоянная схема: [docs/deploy/STABLE_EDGE.md](../../docs/deploy/STABLE_EDGE.md).
- `PASTE-BOT-TOKEN`/`PASTE-TUNNEL-HOST` в создаваемом `.env` — поля, которые вписываются руками на сервере. В git их нет.

## Telegram-эгресс через Cloudflare Worker (tg-proxy)

Yandex Cloud блокирует исходящие к `api.telegram.org`. Входящий webhook идёт через named tunnel. Исходящие вызовы Bot API идут так: приложение → локальный прокси → Cloudflare Worker → Telegram.

Полный порядок (домен, туннель, секрет, что можно присылать в чат): [docs/deploy/STABLE_EDGE.md](../../docs/deploy/STABLE_EDGE.md).

Кратко, уже на ВМ, после деплоя воркера:

```bash
cd /opt/lifeos/repo
sudo LIFEOS_TG_PROXY_WORKER_URL=https://tg-proxy.<account>.workers.dev \
  bash deployments/vps/tg-proxy.sh install
sudo bash deployments/vps/tg-proxy.sh test
```

`test` печатает `ok`, id и username. Токен бота не печатает. Секрет прокси, если он есть, пишется только в `/opt/lifeos/secrets/tg-proxy.env` (`chmod 600`). В `/opt/lifeos/.env` попадает `LIFEOS_HTTP_PROXY`: для Docker это `http://host.docker.internal:8081`, для бинаря на хосте — `http://127.0.0.1:8081`.

Go-клиент при непустом `LIFEOS_HTTP_PROXY` ходит на `http://api.telegram.org`, иначе прокси получил бы CONNECT и не увидел запрос. Отключить: `sudo bash deployments/vps/tg-proxy.sh remove`.
