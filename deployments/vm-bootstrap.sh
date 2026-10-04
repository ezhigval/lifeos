#!/usr/bin/env bash
# LifeOS VM bootstrap — умная установка/обновление окружения.
# Запускать НА ВМ от root:  sudo -s, затем bash vm-bootstrap.sh
# Идемпотентен и безопасен для машины, где LifeOS уже работал:
#   [1] окружение (docker/git) ставится только если отсутствует;
#   [2] старые контейнеры/сервисы lifeos обнаруживаются и останавливаются;
#   [3] БД дампируется ПЕРЕД любыми разрушительными действиями (в /opt/lifeos/backups);
#   [4] репозиторий деплоя пересоздаётся из клонированного main;
#   [5] данные мигрируются: старый docker-том postgres_data переиспользуется как есть,
#       чужой standalone Postgres — вливается dump-ом (pg_restore),
#       fresh volume — pg_restore из последнего бэкапа (если он есть);
#   миграции схемы при старте применяет сам app (embed migrations).
set -uo pipefail

REPO_URL="${LIFEOS_REPO_URL:-git@github.com:ezhigval/lifeos.git}"
BRANCH="${LIFEOS_BRANCH:-main}"
BASE=/opt/lifeos
REPO="$BASE/repo"
CLONE=/root/lifeos-src
BACKUPS="$BASE/backups"
STAMP=$(date +%Y%m%d-%H%M%S)
FAILED=0

step() { echo ""; echo "==> $*"; }
ok()   { echo "    OK: $*"; }
warn() { echo "    !! warning: $*" >&2; FAILED=1; }
die()  { echo "    FATAL: $*" >&2; exit 1; }

run_compose() { # run_compose <compose-file> [args...]
  local cf="$1"; shift
  docker compose -f "$cf" --env-file "$BASE/.env" -p lifeos "$@"
}

# ---------------------------------------------------------------------------
step "[0] Предполётные проверки"
[ "$(id -u)" = 0 ] || die "запускать от root (sudo -s)"
mkdir -p "$BASE" "$BACKUPS"

# ---------------------------------------------------------------------------
step "[1] Окружение: docker, compose, git, ssh (установка ТОЛЬКО если отсутствует)"
if command -v docker >/dev/null 2>&1; then
  ok "docker уже установлен: $(docker --version)"
else
  echo "    docker не найден — устанавливаю через get.docker.com (IPv4) ..."
  apt-get update -y >/dev/null && apt-get install -y ca-certificates curl gnupg >/dev/null
  # Принудительно IPv4: без IPv6-маршрута curl падает на AAAA download.docker.com
  # с "connect (101: Network is unreachable)". DOWNLOAD_URL=https://apt.docker.com —
  # apt-репозиторий Docker работает напрямую, в отличие от Cloudflare-фронта
  # download.docker.com. Если скрипт недоступен — fallback на docker.io из
  # репозитория Ubuntu.
  if curl -4 -fsSL https://get.docker.com -o /tmp/get-docker.sh >/dev/null 2>&1 \
     && { DOWNLOAD_URL="https://apt.docker.com" sh /tmp/get-docker.sh >/dev/null 2>&1 \
          || sh /tmp/get-docker.sh >/dev/null 2>&1; }; then
    ok "docker установлен: $(docker --version)"
  elif apt-get install -y docker.io docker-compose-plugin >/dev/null 2>&1; then
    ok "docker установлен из репозитория Ubuntu: $(docker --version)"
  else
    die "установка docker не удалась"
  fi
fi
systemctl enable --now docker >/dev/null 2>&1 || die "не удалось запустить docker"
if docker compose version >/dev/null 2>&1; then
  ok "docker compose: $(docker compose version --short 2>/dev/null || echo present)"
else
  echo "    плагин docker compose отсутствует — докупаю docker-compose-plugin ..."
  apt-get install -y docker-compose-plugin >/dev/null 2>&1 || \
    die "docker compose plugin missing (apt-get install -y docker-compose-plugin)"
  ok "docker compose установлен"
fi
command -v git >/dev/null 2>&1 || { apt-get install -y git >/dev/null; ok "git установлен"; }
command -v ssh-keyscan >/dev/null 2>&1 || { apt-get install -y openssh-client >/dev/null; ok "openssh-client установлен"; }
command -v openssl >/dev/null 2>&1 || apt-get install -y openssl >/dev/null 2>&1 || true

# ---------------------------------------------------------------------------
step "[2] Обнаружение старых версий LifeOS (контейнеры, сервисы)"
OLD_COMPOSE_FILE=""
OLD_IDS=""
if command -v docker >/dev/null 2>&1; then
  OLD_IDS=$(docker ps -aq --filter "label=com.docker.compose.project=lifeos" 2>/dev/null || true)
  # старые установки под другим project name (например, клон в другой папке):
  EXTRA_IDS=$(docker ps -aq --filter "status=exited" --format '{{.ID}} {{.Names}}' 2>/dev/null \
    | awk '$2 ~ /lifeos/ {print $1}' || true)
  OLD_IDS="$(echo "$OLD_IDS $EXTRA_IDS" | tr ' ' '\n' | grep -v '^$' | sort -u | tr '\n' ' ')"
  if [ -n "$OLD_IDS" ]; then
    echo "    найдены контейнеры проекта lifeos:"
    docker ps -a --filter "label=com.docker.compose.project=lifeos" \
      --format '      {{.Names}}\t{{.Image}}\t{{.Status}}'
    for cid in $OLD_IDS; do
      cf=$(docker inspect -f '{{ index .Config.Labels "com.docker.compose.config-file" }}' "$cid" 2>/dev/null || true)
      [ -n "$cf" ] && OLD_COMPOSE_FILE="$cf" && break
    done
  fi
fi
# сторонние процессы с бинарником lifeos вне docker — останавливаем
STRAY=$(pgrep -af 'lifeos' 2>/dev/null | grep -viE 'grep|pgrep|/root/.ssh|autodeploy' || true)
if [ -n "$STRAY" ]; then
  echo "    сторонние процессы lifeos:"; echo "$STRAY" | sed 's/^/      /'
  echo "$STRAY" | awk '{print $1}' | xargs -r kill 2>/dev/null || true
fi
# systemd-юниты прошлых установок (кроме нашего таймера)
for u in $(systemctl list-units --type=service --all --no-legend 2>/dev/null | awk '{print $1}' | grep -iE '^lifeos' | grep -v '^lifeos-deploy'); do
  echo "    останавливаю старый юнит $u"
  systemctl stop "$u" 2>/dev/null || true
  systemctl disable "$u" 2>/dev/null || true
done
if [ -n "$OLD_IDS" ]; then
  ok "старые контейнеры ПОКА НЕ останавливаю — сначала дамплю живую БД (шаг [3])"
else
  ok "старых running-версий lifeos не обнаружено"
fi

# ---------------------------------------------------------------------------
step "[3] Дамп БД перед любыми изменениями (если кластер жив)"
DUMPED=""
if command -v docker >/dev/null 2>&1; then
  PG=$(docker ps --format '{{.Names}}' 2>/dev/null | grep -iE 'postgres' | head -1 || true)
  if [ -n "$PG" ]; then
    DUMPED="$BACKUPS/pg-preboot-$STAMP.dump"
    echo "    дамплю живой кластер '$PG' -> $DUMPED"
    if docker exec "$PG" pg_dump -U lifeos -d lifeos -F c -f /tmp/preboot.dump >/dev/null 2>&1 && \
       docker cp "$PG":/tmp/preboot.dump "$DUMPED" >/dev/null 2>&1; then
      ok "дамп готов ($(du -h "$DUMPED" | cut -f1))"
    else
      warn "дамп живого кластера не удался (кластер мог быть остановлен на шаге [2]) — использую том ниже"
      DUMPED=""
    fi
  fi
fi
# standalone Postgres в системе (не docker) — тоже дампим на всякий случай
if command -v pg_dumpall >/dev/null 2>&1 && systemctl is-active --quiet postgresql 2>/dev/null; then
  SQLDUMP="$BACKUPS/system-pg-preboot-$STAMP.sql.gz"
  echo "    найден системный postgres — делаю дамп lifeos-базы -> $SQLDUMP"
  if sudo -u postgres pg_dump lifeos 2>/dev/null | gzip > "$SQLDUMP"; then
    ok "системный дамп готов"
  else
    rm -f "$SQLDUMP"; warn "системная база lifeos не найдена или дамп не удался"
  fi
fi


echo "    теперь безопасно останавливаю старый стек..."
if [ -n "$OLD_COMPOSE_FILE" ] && [ -f "$OLD_COMPOSE_FILE" ]; then
  run_compose "$OLD_COMPOSE_FILE" stop postgres app >/dev/null 2>&1 || true
  ok "compose-стек остановлен ($OLD_COMPOSE_FILE), том postgres_data сохранён"
elif [ -n "$OLD_IDS" ]; then
  docker stop $OLD_IDS >/dev/null 2>&1 || true
  ok "контейнеры остановлены по именам"
fi

# ---------------------------------------------------------------------------
step "[4] Deploy-ключ GitHub"
mkdir -p /root/.ssh && chmod 700 /root/.ssh
if [ ! -f /root/.ssh/lifeos_deploy ]; then
  ssh-keygen -t ed25519 -f /root/.ssh/lifeos_deploy -N '' -C "lifeos-deploy@$(hostname)" -q
fi
grep -q 'lifeos_deploy' /root/.ssh/config 2>/dev/null || cat >> /root/.ssh/config <<'CFGEOF'
Host github.com
  IdentityFile /root/.ssh/lifeos_deploy
  IdentitiesOnly yes
CFGEOF
chmod 600 /root/.ssh/config 2>/dev/null || true
ssh-keyscan -t ed25519 github.com >> /root/.ssh/known_hosts 2>/dev/null || true
PUBKEY_FILE=/root/.ssh/lifeos_deploy.pub
[ -s "$PUBKEY_FILE" ] || ssh-keygen -y -f /root/.ssh/lifeos_deploy > "$PUBKEY_FILE"

if ! GIT_SSH_COMMAND="ssh -i /root/.ssh/lifeos_deploy -o IdentitiesOnly=yes" \
     git ls-remote "$REPO_URL" HEAD >/dev/null 2>&1; then
  echo ""
  echo "!!! GitHub ещё НЕ видит deploy-ключ этой машины."
  echo "!!! Добавьте ЭТУ строку как Deploy Key (Settings -> Deploy keys, БЕЗ write access):"
  echo "----------------------------------------------------------------------"
  cat "$PUBKEY_FILE"
  echo "----------------------------------------------------------------------"
  echo "После добавления повторно запустите: bash $0"
  exit 2
fi
ok "deploy key принят (репо читается)"

# ---------------------------------------------------------------------------
step "[5] Клонирование main (чистый источник скриптов и compose)"
export GIT_SSH_COMMAND="ssh -i /root/.ssh/lifeos_deploy -o IdentitiesOnly=yes"
if [ -d "$CLONE/.git" ]; then
  git -C "$CLONE" fetch origin "$BRANCH" >/dev/null && git -C "$CLONE" reset --hard "origin/$BRANCH" >/dev/null
else
  git clone --branch "$BRANCH" "$REPO_URL" "$CLONE" >/dev/null || die "git clone failed"
fi
git config --global --add safe.directory "$CLONE" >/dev/null 2>&1 || true
git config --global --add safe.directory "$REPO" >/dev/null 2>&1 || true
if [ "$REPO" != "$CLONE" ]; then
  rm -rf "$REPO"; mkdir -p "$(dirname "$REPO")"
  cp -a "$CLONE" "$REPO"
fi
ok "код репозитория обновлён ($REPO @ $(git -C "$REPO" rev-parse --short HEAD))"

# ---------------------------------------------------------------------------
step "[6] Инвентаризация данных Postgres и выбор стратегии"
VOL=""
if command -v docker >/dev/null 2>&1; then
  VOL=$(docker volume ls -q --filter "label=com.docker.compose.project=lifeos" 2>/dev/null \
        | grep -E 'postgres_data' | head -1 || true)
  # том от старой установки с другим project name — ищем по имени postgres_data*
  [ -z "$VOL" ] && VOL=$(docker volume ls -q 2>/dev/null | grep -E 'postgres_data' | head -1 || true)
fi
HAS_SYSTEM_PG=false
command -v pg_dump >/dev/null 2>&1 && systemctl is-active --quiet postgresql 2>/dev/null && HAS_SYSTEM_PG=true

# есть ли непустой кластер в найденном томе?
VOL_HAS_DATA=false
if [ -n "$VOL" ]; then
  if [ -n "$DUMPED" ]; then
    VOL_HAS_DATA=true
  elif docker run --rm -v "${VOL}:/d" alpine sh -c 'ls -A /d 2>/dev/null | grep -q .' >/dev/null 2>&1; then
    VOL_HAS_DATA=true
  fi
fi

if [ -n "$VOL" ] && $VOL_HAS_DATA; then
  echo "    стратегия: docker-том '$VOL' с данными -> ПЕРЕИСПОЛЬЗУЕМ как есть (миграции применит app)."
  MIGRATION_NOTE="tom-reuse"
elif $HAS_SYSTEM_PG; then
  echo "    стратегия: docker-тома с данными нет, но есть системный Postgres -> после старта new stack"
  echo "               зальём бэкап дампом в контейнерную базу."
  MIGRATION_NOTE="system-pg-import"
else
  LAST_BACKUP=$(ls -1t "$BACKUPS"/pg-*.dump 2>/dev/null | head -1 || true)
  if [ -n "$LAST_BACKUP" ]; then
    echo "    стратегия: fresh/пустой volume + восстановление из последнего бэкапа $LAST_BACKUP"
    MIGRATION_NOTE="restore-backup"
  else
    echo "    стратегия: чистая установка (данных и бэкапов не найдено)"
    MIGRATION_NOTE="fresh"
  fi
fi

# если системный postgres занимает :5432 и мешает — предупреждаем (порт 5433 у нас, конфликта нет)
$HAS_SYSTEM_PG && echo "    note: системный postgres оставлен включённым; compose использует порт 5433"

# ---------------------------------------------------------------------------
step "[7] Секретный файл .env (существующий НЕ трогаем; при отсутствии — генерация)"
if [ ! -f "$BASE/.env" ]; then
  API_KEY=$(openssl rand -hex 32)
  JWT_SECRET=$(openssl rand -hex 32)
  cat > "$BASE/.env" <<ENVEOF
# --- LifeOS runtime env (сгенерировано vm-bootstrap.sh $(date -u +%FT%TZ)) ---
COMPOSE_PROJECT_NAME=lifeos
LIFEOS_HTTP_ADDR=:8080
LIFEOS_LOG_LEVEL=info
LIFEOS_LOG_FORMAT=text
LIFEOS_TELEGRAM_MODE=polling
TELEGRAM_BOT_TOKEN=
LIFEOS_SEED_TIMEZONE=Europe/Moscow
LIFEOS_SEED_TELEGRAM_ID=0
LIFEOS_SEED_DISPLAY_NAME=Ezhigval
LIFEOS_JWT_SECRET=$JWT_SECRET
LIFEOS_API_KEY=$API_KEY
LIFEOS_JWT_TTL_HOURS=168
LIFEOS_WEBAPP_AUTH_TTL_HOURS=24
LIFEOS_MINIAPP_URL=
LIFEOS_STATIC_DIR=/app/web
LIFEOS_LLM_ENABLED=false
LIFEOS_LLM_AGENT_ENABLED=false
LIFEOS_OTEL_ENABLED=false
ENVEOF
  chmod 600 "$BASE/.env"
  echo "    !! заполните обязательные значения:"
  echo "       TELEGRAM_BOT_TOKEN, LIFEOS_SEED_TELEGRAM_ID, LIFEOS_MINIAPP_URL (опц.)"
  echo "       nano $BASE/.env"
fi
grep -q '^COMPOSE_PROJECT_NAME=' "$BASE/.env" || echo 'COMPOSE_PROJECT_NAME=lifeos' >> "$BASE/.env"
ok ".env на месте"

# ---------------------------------------------------------------------------
step "[8] Сборка и запуск стека (с переиспользованием старого тома, если нужно)"
CF="$REPO/deployments/docker-compose.yml"
OVERRIDE="$BASE/docker-compose.override.yml"
if [ "$MIGRATION_NOTE" = "tom-reuse" ]; then
  # фиксируем старый том за новым проектом
  cat > "$OVERRIDE" <<OVEOF
services:
  postgres:
    volumes:
      - ${VOL}:/var/lib/postgresql/data
volumes:
  ${VOL}:
    external: true
OVEOF
  ok "override: postgres использует существующий том $VOL"
else
  rm -f "$OVERRIDE"
fi
echo "    docker compose up -d --build (первый прогон может занять несколько минут)..."
if [ -f "$OVERRIDE" ]; then
  docker compose -f "$CF" -f "$OVERRIDE" --env-file "$BASE/.env" -p lifeos up -d --build \
    || die "compose up не удался — смотрите docker logs"
else
  docker compose -f "$CF" --env-file "$BASE/.env" -p lifeos up -d --build \
    || die "compose up не удался — смотрите docker logs"
fi

# ---------------------------------------------------------------------------
step "[9] Миграция данных (по стратегии $MIGRATION_NOTE)"
NEWPG=$(docker ps --format '{{.Names}}' 2>/dev/null | grep -E 'lifeos-postgres' | head -1 || true)
case "$MIGRATION_NOTE" in
  tom-reuse|fresh)
    ok "данные уже внутри тома / чистая база — миграции схемы применит app при старте"
    ;;
  restore-backup)
    LAST_BACKUP=$(ls -1t "$BACKUPS"/pg-*.dump 2>/dev/null | head -1 || true)
    if [ -n "$NEWPG" ] && [ -n "$LAST_BACKUP" ]; then
      echo "    восстанавливаю $LAST_BACKUP в $NEWPG ..."
      docker cp "$LAST_BACKUP" "$NEWPG":/tmp/restore.dump && \
      docker exec "$NEWPG" pg_restore -U lifeos -d lifeos --clean --if-exists /tmp/restore.dump \
        >/dev/null 2>&1 && ok "восстановление завершено" \
        || warn "pg_restore вернул ошибки (возможно часть данных уже была) — проверьте вручную"
    fi
    ;;
  system-pg-import)
    IMPORT_SRC=$(ls -1t "$BACKUPS"/system-pg-preboot-*.sql.gz "$BACKUPS"/pg-preboot-*.dump 2>/dev/null | head -1 || true)
    if [ -n "$NEWPG" ] && [ -n "$IMPORT_SRC" ]; then
      echo "    импортирую $IMPORT_SRC в контейнерную базу ..."
      case "$IMPORT_SRC" in
        *.dump)
          docker cp "$IMPORT_SRC" "$NEWPG":/tmp/import.dump && \
          docker exec "$NEWPG" pg_restore -U lifeos -d lifeos --clean --if-exists /tmp/import.dump \
            >/dev/null 2>&1 && ok "dump импортирован" || warn "ошибки pg_restore — проверьте вручную" ;;
        *.sql.gz)
          gunzip -c "$IMPORT_SRC" | docker exec -i "$NEWPG" psql -U lifeos -d lifeos -q >/dev/null 2>&1 \
            && ok "sql-дамп импортирован" || warn "ошибки импорта sql — проверьте вручную" ;;
      esac
    else
      warn "нет источника импорта — пропускаю (данные остались в системном postgres)"
    fi
    ;;
esac

# ---------------------------------------------------------------------------
step "[9.5] Зафиксировать SHA для таймера автодеплоя (не пересобирать зря)"
git -C "$REPO" rev-parse HEAD > /opt/lifeos/.last_deploy_sha
ok "last_deploy_sha=$(cat /opt/lifeos/.last_deploy_sha | head -c 8)"

step "[10] systemd автодеплой"
cp "$REPO/deployments/systemd/lifeos-deploy.service" /etc/systemd/system/
cp "$REPO/deployments/systemd/lifeos-deploy.timer"   /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now lifeos-deploy.timer
ok "таймер автодеплоя активен (каждые 5 мин)"

# ---------------------------------------------------------------------------
step "[11] Health check"
for i in $(seq 1 12); do
  if curl -fsS http://localhost:8080/health >/dev/null 2>&1; then
    echo "    HEALTH OK (попытка $i)"; break
  fi
  sleep 5
  [ "$i" = 12 ] && warn "приложение не отвечает за 60с — journalctl -u lifeos-deploy; docker compose logs app"
done

echo ""
echo "=== ГОТОВО (strategy: $MIGRATION_NOTE) ==="
echo "Бэкапы:     $BACKUPS/"
echo "Логи:       journalctl -u lifeos-deploy -f ; docker compose -f $CF --env-file $BASE/.env logs -f app"
echo "Состояние:  docker compose -f $CF --env-file $BASE/.env ps"
echo "Секреты:    $BASE/.env (chmod 600)"
[ "$FAILED" = 1 ] && echo "!! были предупреждения — см. выше" && exit 1
exit 0
