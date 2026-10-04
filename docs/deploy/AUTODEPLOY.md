# Автодеплой из GitHub (pull-деплой через SSH)

> **v2 (2026-10-04):** скрипт вынесен в `deployments/autodeploy.sh` + unit-файлы
> `deployments/systemd/lifeos-deploy.{service,timer}`. Деплой применяется ТОЛЬКО при
> изменении commit SHA (защита от пересборки каждый цикл). Миграции выполняются внутри
> контейнера app (`/app/lifeos migrate up`) ПЕРЕД рестартом сервиса. ВМ уже работает —
> для первой настройки используйте «Миграция существующей установки» внизу.

 Sandbox агента не имеет доступа к порту 22 ВМ (блокировка на стороне песочницы),
 поэтому деплой инициируется **с самой ВМ** по таймеру systemd — ей ничего не блокировано.

## Архитектура
```
push в main → GitHub ──(git fetch каждые 5 мин)── ВМ: lifeos-deploy.timer
                                                    │ (пропуск, если SHA не изменился)
                                                    ├─ git reset --hard origin/main
                                                    ├─ docker compose build app
                                                    ├─ docker compose up -d postgres && migrate up в app-контейнере
                                                    └─ docker compose up -d app
```
Секреты лежат только на ВМ в `/opt/lifeos/.env` (chmod 600, вне git).

## Быстрый старт (одна команда, выполнить как root на ВМ)

Если ВМ уже клонирует репозиторий где-то ещё — сначала см. «Миграция существующей установки».

```bash
curl -fsSL https://raw.githubusercontent.com/ezhigval/lifeos/main/deployments/autodeploy.sh | bash
```

Для приватного репозитория сначала создайте deploy-ключ (или выполните шаги вручную):

### Шаг 1. Deploy-ключ для приватного репозитория
```bash
ssh-keygen -t ed25519 -f /root/.ssh/lifeos_deploy -N "" -C "lifeos-deploy@vm"
cat /root/.ssh/lifeos_deploy.pub
```
→ GitHub: **ezhigval/lifeos → Settings → Deploy keys → Add deploy key**, галочку *write* ставить НЕ нужно (только чтение).

### Шаг 2. Настройка (или пропустите, если curl выше)
```bash
bash <(curl -fsSL https://raw.githubusercontent.com/ezhigval/lifeos/main/deployments/autodeploy.sh)
```
Скрипт идемпотентен: клонирует репо, создаст `.env`, установит юниты, включит таймер.

### Шаг 3. Заполните секреты
```bash
nano /opt/lifeos/.env     # LIFEOS_TELEGRAM_BOT_TOKEN, LIFEOS_JWT_SECRET, LIFEOS_API_KEY и т.д.
                          # (см. YANDEX_CLOUD_SECRETS.md)
```

### Шаг 4. Первый деплой вручную + проверка
```bash
systemctl start lifeos-deploy.service
journalctl -u lifeos-deploy -n 80 --no-pager
curl -fsS http://localhost:8080/health
docker compose -f /opt/lifeos/repo/deployments/docker-compose.yml ps
```

Дальше всё автоматически: merge в `main` → в течение ~5 минут на ВМ поднимется новая версия.

## Ручной деплой (немедленно)
```bash
systemctl start lifeos-deploy.service
journalctl -u lifeos-deploy -n 50 --no-pager
curl -fsS http://localhost:8080/health
```

## Диагностика
```bash
systemctl list-timers lifeos-deploy.timer        # следующий запуск
cat /opt/lifeos/.last_deploy_sha                 # задеплоенный коммит
git -C /opt/lifeos/repo log --oneline -3         # что в клоне
journalctl -u lifeos-deploy --since "1 hour ago" --no-pager
docker logs lifeos-app-1 --tail 50               # имя контейнера = lifeos-app (compose project lifeos)
```

## Миграция существующей установки (ВМ уже работает)
Если приложение уже запущено из другой директории (например, `/home/smailikin70/lifeos`):
```bash
# 1. Остановите старый способ запуска (если это systemd-сервис или docker compose из другого места)
#    Данные БД останутся в docker volume postgres_data — они не потеряются,
#    ЕСЛИ volume принадлежит тому же compose-проекту. Проверьте:
docker volume ls | grep postgres_data
# 2. Выполните autodeploy.sh — он использует compose-проект "lifeos"
#    (project name задаётся COMPOSE_PROJECT_NAME=lifeos в /opt/lifeos/.env, см. скрипт).
#    Если старый volume называется иначе, переименуйте:
docker volume create lifeos_postgres_data
docker run --rm -v СТАРЫЙ_volume:/from -v lifeos_postgres_data:/to alpine \
  sh -c 'cp -a /from/. /to/'
# 3. Первый запуск: миграции применятся автоматически (migrate up перед рестартом app).
```
Если порт 8080 занят старым контейнером — сначала `docker stop` его.

## Если репозиторий приватный (SSH deploy-ключ)
```bash
ssh-keygen -t ed25519 -f /root/.ssh/lifeos_deploy -N "" -C "lifeos-deploy@vm"
cat /root/.ssh/lifeos_deploy.pub   # → GitHub Settings → Deploy keys (read-only достаточно)
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


## Полный умный установщик (рекомендуемый путь)

`deployments/vm-bootstrap.sh` — идемпотентная установка/обновление «одной командой».
Перед любыми разрушительными действиями сам находит и безопасно снимает старые версии:

| Шаг | Что делает |
|-----|------------|
| [1] | Окружение: docker / compose plugin / git / openssh ставятся **только если отсутствуют** (повторный запуск ничего не переустанавливает) |
| [2] | Обнаружение старых версий lifeos: контейнеры compose-проекта `lifeos` (в т.ч. exited под другим project name), сторонние процессы с бинарником `lifeos`, systemd-юниты `lifeos*` (кроме таймера) — останавливаются |
| [3] | **Дамп БД ПЕРЕД остановкой**: живой docker-postgres → `pg_dump -F c` в `/opt/lifeos/backups/pg-preboot-*.dump`; системный (standalone) postgres → sql.gz-дамп базы `lifeos`. Только после дампа старый стек останавливается (том `postgres_data` сохраняется) |
| [4–5] | Deploy-ключ GitHub (exit 2 + печать pubkey, если ключ не добавлен), клон `main` в `/root/lifeos-src` → `/opt/lifeos/repo` |
| [6] | Инвентаризация данных и выбор стратегии: `tom-reuse` (найдён непустой docker-том — переиспользуется как есть) / `system-pg-import` (данные в системном postgres — зальём дампом) / `restore-backup` (fresh volume + последний бэкап) / `fresh` |
| [7] | `.env`: существующий **не трогается**; при отсутствии генерируется со случайными API_KEY/JWT_SECRET |
| [8–9] | `compose up -d --build` (при tom-reuse — override на старый том), затем миграция данных по стратегии; миграции схемы применяет app (`migrate up`) |
| [9.5–11] | Фиксация `.last_deploy_sha` (таймер не будет пересобирать зря), установка systemd-юнитов, health-check с ретраями до 60с |

```bash
# на ВМ, от root (raw-доступ к приватному репо не работает — скопируйте
# deployments/vm-bootstrap.sh через scp/буфер обмена своего ssh-клиента):
sudo bash vm-bootstrap.sh
```

Если GitHub ещё не знает deploy-ключ машины, скрипт остановится (exit 2) и покажет
строку pubkey — добавьте её в **Settings → Deploy keys** и запустите скрипт повторно.

### Снос всего и чистая переустановка

```bash
# остановить таймер и приложение, удалить контейнеры и образы
systemctl stop lifeos-deploy.timer lifeos-deploy.service || true
docker compose -f /opt/lifeos/repo/deployments/docker-compose.yml --env-file /opt/lifeos/.env down -v --rmi all || true
rm -rf /opt/lifeos/repo /etc/systemd/system/lifeos-deploy.{service,timer}
systemctl daemon-reload
# ДАННЫЕ: том postgres_data удалён вместе с -v выше. Если БД нужна — НЕ используйте down -v,
# а только `down` (том останется и подключится к новому кластеру).
# затем заново: sudo bash vm-bootstrap.sh
```
