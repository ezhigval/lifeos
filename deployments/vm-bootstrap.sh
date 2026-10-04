#!/usr/bin/env bash
# LifeOS VM bootstrap — полный чистый установочный скрипт.
# Запускать НА ВМ от root:  sudo -s, затем bash vm-bootstrap.sh
# Идемпотентен: повторный запуск безопасен (секреты в /opt/lifeos/.env не перезаписываются).
set -euo pipefail

REPO_URL="${LIFEOS_REPO_URL:-git@github.com:ezhigval/lifeos.git}"
BRANCH="${LIFEOS_BRANCH:-main}"
BASE=/opt/lifeos
REPO="$BASE/repo"

echo "==> [0] Проверка окружения"
[ "$(id -u)" = 0 ] || { echo "run as root"; exit 1; }

echo "==> [1] Docker (если нет — установка с get.docker.com)"
if ! command -v docker >/dev/null 2>&1; then
  apt-get update -y && apt-get install -y ca-certificates curl gnupg
  curl -fsSL https://get.docker.com | sh
fi
systemctl enable --now docker
docker compose version >/dev/null || { echo "docker compose plugin missing"; exit 1; }

echo "==> [2] git + ssh"
command -v git >/dev/null || apt-get install -y git
command -v ssh-keyscan >/dev/null || apt-get install -y openssh-client

echo "==> [3] Deploy-ключ для приватного репо GitHub"
mkdir -p /root/.ssh && chmod 700 /root/.ssh
if [ ! -f /root/.ssh/lifeos_deploy ]; then
  ssh-keygen -t ed25519 -f /root/.ssh/lifeos_deploy -N '' -C "lifeos-deploy@$(hostname)" -q
fi
cat >> /root/.ssh/config <<EOF || true
Host github.com
  IdentityFile /root/.ssh/lifeos_deploy
  IdentitiesOnly yes
EOF
chmod 600 /root/.ssh/config 2>/dev/null || true
ssh-keyscan -t ed25519 github.com >> /root/.ssh/known_hosts 2>/dev/null || true

PUBKEY_FILE=/root/.ssh/lifeos_deploy.pub
if [ ! -s "$PUBKEY_FILE" ]; then ssh-keygen -y -f /root/.ssh/lifeos_deploy > "$PUBKEY_FILE"; fi

if ! GIT_SSH_COMMAND="ssh -i /root/.ssh/lifeos_deploy -o IdentitiesOnly=yes" \
     git ls-remote "$REPO_URL" HEAD >/dev/null 2>&1; then
  echo ""
  echo "!!! GitHub ещё НЕ видит deploy-ключ этой машины."
  echo "!!! Добавьте ЭТУ строку как Deploy Key (Settings -> Deploy keys, без write access):"
  echo "----------------------------------------------------------------------"
  cat "$PUBKEY_FILE"
  echo "----------------------------------------------------------------------"
  echo "После добавления повторно запустите: bash $0"
  exit 2
fi
echo "==> Deploy key OK (репо читается)"

echo "==> [4] Клонирование/обновление репозитория"
if [ ! -d "$REPO/.git" ]; then
  git clone --branch "$BRANCH" "$REPO_URL" "$REPO"
else
  git -C "$REPO" fetch origin "$BRANCH" && git -C "$REPO" reset --hard "origin/$BRANCH"
fi
git config --global --add safe.directory "$REPO" || true

echo "==> [5] .env из примера (секреты существующие не трогаем)"
if [ ! -f "$BASE/.env" ]; then
  cp "$REPO/.env.example" "$BASE/.env"
  chmod 600 "$BASE/.env"
  # автоподстановка криптостойких секретов, если остались CHANGE_ME
  grep -q '^LIFEOS_API_KEY=CHANGE' "$BASE/.env" && \
    sed -i "s|^LIFEOS_API_KEY=.*|LIFEOS_API_KEY=$(openssl rand -hex 32)|" "$BASE/.env"
  grep -q '^LIFEOS_JWT_SECRET=CHANGE' "$BASE/.env" && \
    sed -i "s|^LIFEOS_JWT_SECRET=.*|LIFEOS_JWT_SECRET=$(openssl rand -hex 32)|" "$BASE/.env"
  grep -q '^LIFEOS_LEARNING_SALT=CHANGE' "$BASE/.env" && \
    sed -i "s|^LIFEOS_LEARNING_SALT=.*|LIFEOS_LEARNING_SALT=$(openssl rand -hex 16)|" "$BASE/.env"
  echo "!! ВАЖНО: заполните TELEGRAM_BOT_TOKEN и LIFEOS_SEED_TELEGRAM_ID в $BASE/.env"
  echo "   nano $BASE/.env"
fi
grep -q '^COMPOSE_PROJECT_NAME=' "$BASE/.env" || echo 'COMPOSE_PROJECT_NAME=lifeos' >> "$BASE/.env"

echo "==> [6] systemd юниты автодеплоя (из репозитория)"
cp "$REPO/deployments/systemd/lifeos-deploy.service" /etc/systemd/system/
cp "$REPO/deployments/systemd/lifeos-deploy.timer"   /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now lifeos-deploy.timer

echo "==> [7] Первый деплой"
systemctl start lifeos-deploy.service
journalctl -u lifeos-deploy -n 40 --no-pager || true

echo "==> [8] Health check"
sleep 2
curl -fsS http://localhost:8080/health && echo " -> HEALTH OK" || { echo "app не отвечает — смотрите journalctl -u lifeos-deploy"; exit 1; }

echo ""
echo "=== ГОТОВО ==="
echo "Автодеплой: таймер каждые 5 мин тянет main; пересборка только при новом коммите."
echo "Логи:       journalctl -u lifeos-deploy -f"
echo "Состояние:  docker compose -f $REPO/deployments/docker-compose.yml --env-file $BASE/.env ps"
echo "Секреты:    $BASE/.env (chmod 600)"
