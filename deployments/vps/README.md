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
