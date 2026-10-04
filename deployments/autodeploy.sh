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
if ! command -v docker >/dev/null 2>&1; then
  echo "==> Installing Docker"
  curl -fsSL https://get.docker.com | sh
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

# 5. Юниты systemd (берём из клонированного репо)
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
After that the timer pulls main every 5 minutes (rebuild only on new commit).
"""
