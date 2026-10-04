# Автодеплой из GitHub (pull-деплой через SSH)

 Sandbox агента не имеет доступа к порту 22 ВМ (блокировка на стороне песочницы),
 поэтому деплой инициируется **с самой ВМ** по таймеру systemd — ей ничего не блокировано.

## Архитектура
```
push в main → GitHub ──(git pull каждые 5 мин)── ВМ: lifeos-deploy.timer
                                                 ├─ git pull --ff-only
                                                 ├─ docker compose build + up -d
                                                 └─ миграции (app стартует с migrate)
```
Секреты лежат только на ВМ в `/opt/lifeos/.env` (chmod 600, вне git).

## Одноразовая настройка на ВМ (выполнить как root через ваш SSH/com1)

```bash
set -euo pipefail

# 1. Директории и репозиторий
mkdir -p /opt/lifeos/backups
cd /opt/lifeos
git clone git@github.com:ezhigval/lifeos.git repo   # нужен deploy-ключ (см. ниже)
# или HTTPS без ключа:
# git clone https://github.com/ezhigval/lifeos.git repo

# 2. Секреты (один раз, вручную; см. YANDEX_CLOUD_SECRETS.md)
cp repo/.env.example /opt/lifeos/.env
chmod 600 /opt/lifeos/.env
nano /opt/lifeos/.env        # боевые значения: bot token, JWT, API keys

# 3. Docker (если ещё нет)
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker

# 4. Продакшен override (без секретов, но вне git — не коммитить)
cat > /opt/lifeos/repo/deployments/docker-compose.override.yml <<'OVR'
services:
  postgres:
    ports: []          # не светить 5432 наружу
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/pgpass
    volumes:
      - pgpass:/run/secrets:ro
volumes:
  pgpass:
OVR
echo "lifeos-prod-pg-pass" > /dev/null  # пароль БД задайте сами и пропишите в .env/app

# 5. Сервис автодеплоя
cat > /etc/systemd/system/lifeos-deploy.service <<'SVC'
[Unit]
Description=LifeOS pull-deploy
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
WorkingDirectory=/opt/lifeos/repo
ExecStart=/usr/bin/git fetch origin main
ExecStart=/usr/bin/git checkout main
ExecStart=/usr/bin/git reset --hard origin/main
ExecStart=/usr/bin/docker compose -f deployments/docker-compose.yml build app
ExecStart=/usr/bin/docker compose -f deployments/docker-compose.yml up -d
ExecStartPre=-/usr/bin/docker exec lifeos-repo-app-1 /app/lifeos migrate up
TimeoutStartSec=1800
SVC

cat > /etc/systemd/system/lifeos-deploy.timer <<'TMR'
[Unit]
Description=Run LifeOS deploy every 5 minutes

[Timer]
OnBootSec=2min
OnUnitActiveSec=5min
RandomizedDelaySec=30

[Install]
WantedBy=timers.target
TMR

systemctl daemon-reload
systemctl enable --now lifeos-deploy.timer
```

> Имя контейнера app может отличаться (`docker ps`); проверьте `ExecStartPre` —
> либо уберите его, если миграции применяются при старте образа.

## Ручной деплой (немедленно)
```bash
systemctl start lifeos-deploy.service
journalctl -u lifeos-deploy -n 50 --no-pager
curl -fsS http://localhost:8080/health
```

## Если репозиторий приватный (SSH deploy-ключ)
```bash
ssh-keygen -t ed25519 -f /root/.ssh/lifeos_deploy -N "" -C "lifeos-deploy@vm"
cat /root/.ssh/lifeos_deploy.pub   # → GitHub Settings → Deploy keys (check "write")
git config --global --add safe.directory /opt/lifeos/repo
```
В `/root/.ssh/config`:
```
Host github.com
    IdentityFile /root/.ssh/lifeos_deploy
```

## Вариант B: push-деплой из GitHub Actions (когда порт 22 ВМ откроют)
Создать secret `VM_SSH_KEY` (приватный deploy-ключ) и workflow `.github/workflows/deploy.yml`:
```yaml
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Deploy over SSH
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.VM_HOST }}
          username: ${{ secrets.VM_USER }}
          key: ${{ secrets.VM_SSH_KEY }}
          script: |
            cd /opt/lifeos/repo
            git pull --ff-only
            docker compose -f deployments/docker-compose.yml up -d --build
```
Пока port 22 закрыт для внешних пушеров — используйте вариант A (timer на ВМ).

## Откат
```bash
cd /opt/lifeos/repo
git checkout <prev-tag-or-commit>
docker compose -f deployments/docker-compose.yml up -d --build
```
