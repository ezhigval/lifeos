#!/bin/bash
# Cloud-init user-data для первого запуска VDS Timeweb (Ubuntu 24.04, Нидерланды).
# Вставить файл целиком в поле User data при создании сервера.
#
# Скрипт ставит Docker, открывает 22/80/443, заводит /opt/lifeos и генерирует
# секреты в /opt/lifeos/.env. Токен бота и почту для сертификата не выдумывает.
# Приложение и Caddy не запускает: A-запись на этот IP ещё может не существовать,
# а Let's Encrypt на пустом домене только сожжёт лимит.
#
# В файрволе панели Timeweb тоже должны быть открыты TCP 22, 80 и 443.
# Повторный запуск не затирает .env:
#   sudo /usr/local/sbin/lifeos-bootstrap
set -euo pipefail
set +x

LIFEOS_DOMAIN="${LIFEOS_DOMAIN:-local-ai-assist.ru}"
LIFEOS_REPO_URL="${LIFEOS_REPO_URL:-https://github.com/ezhigval/lifeos.git}"
LIFEOS_BRANCH="${LIFEOS_BRANCH:-main}"
BASE="${LIFEOS_BASE:-/opt/lifeos}"
APPLY="${LIFEOS_BOOTSTRAP_APPLY:-1}"

log() { echo "[lifeos-bootstrap] $*"; }

rand_hex() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
    return
  fi
  python3 -c 'import secrets; print(secrets.token_hex(32))'
}

require_root() {
  if [[ "$(id -u)" -ne 0 ]]; then
    echo "нужен root" >&2
    exit 1
  fi
}

install_packages() {
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -y
  apt-get install -y --no-install-recommends ca-certificates curl git ufw openssl python3
  # Пакеты Ubuntu, без get.docker.com: на чистой VPS этот скрипт не должен
  # зависеть от ещё одного репозитория.
  apt-get install -y docker.io docker-compose-v2
  systemctl enable --now docker
}

configure_docker_logs() {
  install -d -m 755 /etc/docker
  local cfg=/etc/docker/daemon.json
  if [[ -f "$cfg" ]]; then
    return 0
  fi
  cat > "$cfg" <<'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "live-restore": true
}
EOF
  systemctl restart docker
}

configure_swap() {
  if ! swapon --show | grep -q '/swapfile'; then
    fallocate -l 1G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
  fi
  grep -q '/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
  echo 'vm.swappiness=10' > /etc/sysctl.d/99-lifeos-swap.conf
  sysctl -w vm.swappiness=10 >/dev/null
}

configure_ufw() {
  ufw default deny incoming
  ufw default allow outgoing
  ufw allow 22/tcp
  ufw allow 80/tcp
  ufw allow 443/tcp
  ufw --force enable
}

clone_repo() {
  install -d -m 755 "$BASE"
  if [[ -d "$BASE/repo/.git" ]]; then
    log "repo already present"
    return 0
  fi
  git clone --branch "$LIFEOS_BRANCH" --depth 1 "$LIFEOS_REPO_URL" "$BASE/repo"
  git config --global --add safe.directory "$BASE/repo" || true
}

write_env() {
  if [[ -f "$BASE/.env" ]]; then
    log ".env already exists"
    return 0
  fi
  local jwt pg api wh
  jwt="$(rand_hex)"
  pg="$(rand_hex)"
  api="$(rand_hex)"
  wh="$(rand_hex)"
  python3 - "$BASE/.env" "$LIFEOS_DOMAIN" "$jwt" "$pg" "$api" "$wh" <<'PY'
import os, sys
path, domain, jwt, pg, api, wh = sys.argv[1:]
text = f"""# Создан cloud-init Timeweb. Токен и почта пустые нарочно.
# chmod 600. В git не класть. LIFEOS_HTTP_PROXY здесь нет: до Telegram ход прямой.

LIFEOS_DOMAIN={domain}
CADDY_ACME_EMAIL=

POSTGRES_USER=lifeos
POSTGRES_PASSWORD={pg}
POSTGRES_DB=lifeos

TELEGRAM_BOT_TOKEN=
LIFEOS_TELEGRAM_MODE=webhook
LIFEOS_TELEGRAM_WEBHOOK_SECRET={wh}
LIFEOS_SEED_TELEGRAM_ID=0
LIFEOS_SEED_TIMEZONE=Europe/Moscow

LIFEOS_JWT_SECRET={jwt}
LIFEOS_API_KEY={api}
LIFEOS_JWT_TTL_HOURS=168
LIFEOS_WEBAPP_AUTH_TTL_HOURS=24

LIFEOS_LOG_LEVEL=info
LIFEOS_LOG_FORMAT=text
LIFEOS_LLM_ENABLED=false
COMPOSE_PROJECT_NAME=lifeos
"""
fd = os.open(path, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
os.write(fd, text.encode())
os.close(fd)
PY
  chmod 600 "$BASE/.env"
}

write_net_check() {
  install -d -m 755 "$BASE/bin" "$BASE/backups" "$BASE/secrets"
  chmod 700 "$BASE/secrets"
  cat > "$BASE/bin/net-check.sh" <<'EOF'
#!/bin/bash
# Доступность с этой VPS. Код 000 — таймаут или обрыв, любой другой HTTP значит маршрут есть.
set -u
fail=0
check() {
  local url="$1" code
  code="$(curl -4 -sS -o /dev/null -w '%{http_code}' --max-time 10 "$url" || true)"
  if [[ -z "$code" || "$code" == "000" ]]; then
    echo "FAIL $url"
    fail=1
  else
    echo "ok   $url ($code)"
  fi
}
check https://api.telegram.org
check https://github.com
check https://dl-cdn.alpinelinux.org
exit "$fail"
EOF
  chmod 755 "$BASE/bin/net-check.sh"
}

write_next_steps() {
  cat > "$BASE/NEXT.txt" <<EOF
Дальше, по SSH:

1. Проверь сеть (любой FAIL — эту локацию не использовать):
   $BASE/bin/net-check.sh

2. Впиши в $BASE/.env две пустые строки, остальное уже сгенерировано:
   TELEGRAM_BOT_TOKEN
   CADDY_ACME_EMAIL

3. A-запись ${LIFEOS_DOMAIN} на публичный IPv4 этой VPS.
   В панели Timeweb открой TCP 22, 80, 443.
   Оранжевое облако Cloudflare для выпуска сертификата должно быть выключено.

4. Стек поднимается отдельно, когда токен и DNS готовы.
   Этот скрипт приложение не запускает.
EOF
  if [[ "$APPLY" == "1" ]]; then
    cat > /etc/update-motd.d/99-lifeos <<EOF
#!/bin/sh
echo "LifeOS: заполни TELEGRAM_BOT_TOKEN и CADDY_ACME_EMAIL в $BASE/.env"
echo "Сеть: $BASE/bin/net-check.sh"
echo "Шаги: $BASE/NEXT.txt"
EOF
    chmod 755 /etc/update-motd.d/99-lifeos
  fi
}

keep_rerunnable_copy() {
  if [[ -f "${0:-}" && "${0}" != "/usr/local/sbin/lifeos-bootstrap" ]]; then
    install -m 755 "$0" /usr/local/sbin/lifeos-bootstrap
  fi
}

main() {
  if [[ "$APPLY" == "1" ]]; then
    require_root
    keep_rerunnable_copy
    log "packages"
    install_packages
    log "docker logs"
    configure_docker_logs
    log "swap"
    configure_swap
    log "ufw"
    configure_ufw
    timedatectl set-timezone UTC || true
    log "clone"
    clone_repo
  else
    log "dry run, system packages skipped"
    install -d -m 755 "$BASE"
  fi
  write_env
  write_net_check
  write_next_steps
  log "ready ($BASE)"
}

main "$@"
