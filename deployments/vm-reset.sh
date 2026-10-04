#!/usr/bin/env bash
# LifeOS VM reset — сносит ВСЁ связанное с приложением и ставит окружение заново.
# Запускать НА ВМ от root:  sudo bash vm-reset.sh
# Данные Postgres сохраняются в бэкап перед сносом (/opt/lifeos/backups/).
set -uo pipefail

BASE=/opt/lifeos
REPO="$BASE/repo"
STAMP=$(date +%Y%m%d-%H%M%S)

echo "==> [1/4] Остановка автодеплоя и контейнеров (том postgres_data НЕ удаляем)"
systemctl stop lifeos-deploy.timer lifeos-deploy.service 2>/dev/null || true
systemctl disable lifeos-deploy.timer 2>/dev/null || true
if command -v docker >/dev/null 2>&1; then
  # бэкап БД перед удалением тома (если кластер жив)
  PG=$(docker ps --format '{{.Names}}' 2>/dev/null | grep -iE 'postgres' | head -1 || true)
  if [ -n "$PG" ]; then
    echo "==> Бэкап БД -> $BASE/backups/pg-$STAMP.dump"
    mkdir -p "$BASE/backups"
    docker exec "$PG" pg_dump -U lifeos -d lifeos -F c -f "/tmp/pg-$STAMP.dump" 2>/dev/null && \
      docker cp "$PG":"/tmp/pg-$STAMP.dump" "$BASE/backups/" 2>/dev/null || \
      echo "!! бэкап не удался (контейнер недоступен) — продолжение с сохранением тома"
  fi
  CF="$REPO/deployments/docker-compose.yml"
  if [ -f "$CF" ]; then
    docker compose -f "$CF" --env-file "$BASE/.env" -p lifeos down 2>/dev/null || true
  else
    docker ps -aq --filter "label=com.docker.compose.project=lifeos" 2>/dev/null | xargs -r docker stop 2>/dev/null || true
  fi
fi

echo "==> [2/4] Удаление юнитов, репозитория (НЕ трогаем .env и бэкапы)"
rm -f /etc/systemd/system/lifeos-deploy.service /etc/systemd/system/lifeos-deploy.timer
systemctl daemon-reload
rm -rf "$REPO"
docker image prune -af >/dev/null 2>&1 || true

echo "==> [3/4] Чистая переустановка через vm-bootstrap.sh"
BOOTSTRAP_FROM="${LIFEOS_BOOTSTRAP:-}"
if [ -n "$BOOTSTRAP_FROM" ]; then
  bash "$BOOTSTRAP_FROM"
else
  # bootstrap ещё не склонирован — скачиваем его из репо НЕВОЗМОЖНО без ключа;
  # используем копию рядом со скриптом (vm-reset.sh лежит в deployments/) или inline-подсказку.
  CANDIDATE="$(dirname "$0")/vm-bootstrap.sh"
  if [ -f "$CANDIDATE" ]; then
    bash "$CANDIDATE"
  else
    echo ""
    echo "!!! Не найден vm-bootstrap.sh. Положите его рядом с этим скриптом на ВМ"
    echo "!!! (файл deployments/vm-bootstrap.sh из репозитория ezhigval/lifeos, ветка main)"
    echo "!!! и запустите повторно: sudo bash vm-reset.sh"
    exit 1
  fi
fi

echo "==> [4/4] Восстановление данных из бэкапа (если есть новый кластер и старый том)"
LAST_BACKUP=$(ls -1t "$BASE/backups"/pg-*.dump 2>/dev/null | head -1 || true)
if [ -n "$LAST_BACKUP" ]; then
  echo "Бэкап лежит здесь: $LAST_BACKUP"
  echo "Восстановить при необходимости:"
  echo "  docker exec -i lifeos-postgres-1 pg_restore -U lifeos -d lifeos --clean < '$LAST_BACKUP'"
fi

echo ""
echo "=== RESET+REINSTALL завершён ==="
curl -fsS http://localhost:8080/health && echo " <- HEALTH OK" || echo "!! app не отвечает — journalctl -u lifeos-deploy"
