#!/usr/bin/env bash
# LifeOS pull-deploy bootstrap. Идемпотентен: можно запускать повторно.
# Usage (на ВМ, как root):  curl -fsSL .../deployments/autodeploy.sh | bash
set -euo pipefail

REPO_URL="${LIFEOS_REPO_URL:-git@github.com:ezhigval/lifeos.git}"
BRANCH="${LIFEOS_BRANCH:-main}"
BASE=/opt/lifeos
REPO="$BASE/repo"

echo "==> LifeOS autodeploy setup (branch=$BRANCH)"

# 1. Docker
install_docker() {
  # Принудительно IPv4: на ВМ без IPv6-маршрута get.docker.com падает с
  # "connect (101: Network is unreachable)" из-за AAAA-записи download.docker.com.
  # mirror через переменную окружения DOWNLOAD_URL: apt-репозиторий Docker
  # (https://apt.docker.com) работает напрямую, в отличие от Cloudflare-фронта
  # download.docker.com. Скрипт подхватывает DOWNLOAD_URL, если он задан.
  if curl -4 -fsSL https://get.docker.com -o /tmp/get-docker.sh; then
    if DOWNLOAD_URL="https://apt.docker.com" sh /tmp/get-docker.sh; then return 0; fi
    if sh /tmp/get-docker.sh; then return 0; fi
  fi
  echo "!! get.docker.com недоступен — ставим docker из репозитория Ubuntu" >&2
  apt-get update -y
  apt-get install -y docker.io docker-compose-plugin
}

if ! command -v docker >/dev/null 2>&1; then
  echo "==> Installing Docker"
  install_docker
fi
systemctl enable --now docker

# 2. Директории
mkdir -p "$BASE/backups"

# 3. Репозиторий
if [ ! -d "$REPO/.git" ]; then
  echo "==> Cloning $REPO_URL"
  git clone --branch "$BRANCH" "$REPO_URL" "$REPO"
else
  echo "==> Repo already cloned"
fi
git config --global --add safe.directory "$REPO" || true

# 4. .env из примера (секреты не трогаем, если уже есть)
if [ ! -f "$BASE/.env" ]; then
  cp "$REPO/.env.example" "$BASE/.env"
  chmod 600 "$BASE/.env"
  # compose project name — стабильные имена контейнеров lifeos-app-1 / lifeos-postgres-1
  grep -q '^COMPOSE_PROJECT_NAME=' "$BASE/.env" || echo 'COMPOSE_PROJECT_NAME=lifeos' >> "$BASE/.env"
  echo "!! Created $BASE/.env from example — FILL IN SECRETS before first deploy:"
  echo "   nano $BASE/.env"
else
  grep -q '^COMPOSE_PROJECT_NAME=' "$BASE/.env" || echo 'COMPOSE_PROJECT_NAME=lifeos' >> "$BASE/.env"
fi

# 5. Pull-deploy script, then the timer that runs it.
install -d -m 755 "$BASE/bin"
install -m 755 "$REPO/deployments/vm-pull-deploy.sh" "$BASE/bin/vm-pull-deploy.sh"
if [[ -f "$REPO/deployments/nginx/lifeos.conf" ]]; then
  install -m 644 "$REPO/deployments/nginx/lifeos.conf" "$BASE/bin/nginx-lifeos.conf"
fi
cp "$REPO/deployments/systemd/lifeos-deploy.service" /etc/systemd/system/
cp "$REPO/deployments/systemd/lifeos-deploy.timer"   /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now lifeos-deploy.timer

echo """
==> Done. Next steps:
  1) Fill secrets:      nano $BASE/.env
  2) First deploy now:  systemctl start lifeos-deploy.service
                        journalctl -u lifeos-deploy -n 80 --no-pager
  3) Verify:            curl -fsS http://localhost:8080/health
After that the timer pulls main once a day at 04:15 UTC (rebuild only on new commit).
"""
