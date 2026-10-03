#!/usr/bin/env bash
# LifeOS single-shot VPS installer (Ubuntu 22.04/24.04, e.g. Yandex Cloud Compute).
# Run as root on a fresh VM:  sudo bash install.sh
# Idempotent: safe to re-run for updates.
set -euo pipefail

APP_DIR=/opt/lifeos
GO_VERSION=1.25.5
REPO_URL="${REPO_URL:-}"   # set before running, or store in /opt/lifeos/repo_url

log() { echo "[install] $*"; }
export DEBIAN_FRONTEND=noninteractive

apt-get update -y
apt-get install -y git curl ca-certificates gnupg postgresql nginx ufw rsync unzip

# --- PostgreSQL role/db -----------------------------------------------------
sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='lifeos'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE ROLE lifeos LOGIN PASSWORD 'change-me-pg-pass';"
sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='lifeos'" | grep -q 1 || \
  sudo -u postgres createdb -O lifeos lifeos

# --- Go ---------------------------------------------------------------------
if ! /usr/local/go/bin/go version >/dev/null 2>&1; then
  log "Installing Go ${GO_VERSION}"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz
  rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tgz && rm /tmp/go.tgz
fi

# --- cloudflared --------------------------------------------------------------
if ! command -v cloudflared >/dev/null 2>&1; then
  log "Installing cloudflared"
  curl -fsSL https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb -o /tmp/cfd.deb
  apt-get install -y /tmp/cfd.deb && rm /tmp/cfd.deb
fi

# --- source -------------------------------------------------------------------
mkdir -p "$APP_DIR"
if [ -d "$APP_DIR/src/.git" ]; then
  [ -n "$REPO_URL" ] || REPO_URL="$(cat "$APP_DIR/repo_url" 2>/dev/null || true)"
  git -C "$APP_DIR/src" remote set-url origin "$REPO_URL" 2>/dev/null || true
  git -C "$APP_DIR/src" fetch origin && git -C "$APP_DIR/src" reset --hard origin/main
else
  [ -n "$REPO_URL" ] || { echo "REPO_URL not set (env or /opt/lifeos/repo_url)"; exit 1; }
  echo "$REPO_URL" > "$APP_DIR/repo_url"
  git clone "$REPO_URL" "$APP_DIR/src"
fi
cd "$APP_DIR/src"

# --- build --------------------------------------------------------------------
export PATH=/usr/local/go/bin:$PATH
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$APP_DIR/bin/lifeos" ./cmd/lifeos
if command -v node >/dev/null 2>&1; then
  (cd web/miniapp && npm ci --no-audit --no-fund && npm run build)
else
  log "node absent — using committed web/miniapp/dist"
fi
cp -r migrations "$APP_DIR/migrations"

# --- env (create once; never overwrite secrets) --------------------------------
if [ ! -f "$APP_DIR/lifeos.env" ]; then
  JWT=$(openssl rand -hex 32); AK=$(openssl rand -hex 16)
  cat > "$APP_DIR/lifeos.env" <<ENV
LIFEOS_DATABASE_URL=postgres://lifeos:change-me-pg-pass@/lifeos?host=/var/run/postgresql&sslmode=disable
LIFEOS_HTTP_ADDR=:8080
LIFEOS_LOG_LEVEL=info
LIFEOS_LOG_FORMAT=json
LIFEOS_TELEGRAM_MODE=polling
LIFEOS_TELEGRAM_BOT_TOKEN=${TG_TOKEN:-PASTE-BOT-TOKEN}
LIFEOS_SEED_TIMEZONE=Europe/Moscow
LIFEOS_JWT_SECRET=${JWT}
LIFEOS_API_KEY=${AK}
LIFEOS_WEBAPP_AUTH_TTL_HOURS=24
LIFEOS_STATIC_DIR=${APP_DIR}/src/web/miniapp/dist
LIFEOS_MINIAPP_URL=https://PASTE-TUNNEL-HOST/app/
LIFEOS_LLM_ENABLED=false
LIFEOS_STT_ENABLED=false
LIFEOS_VISION_ENABLED=false
ENV
  chmod 600 "$APP_DIR/lifeos.env"
  log "Created $APP_DIR/lifeos.env — PASTE-BOT-TOKEN и PASTE-TUNNEL-HOST обязательны."
fi

# --- systemd: app ---------------------------------------------------------------
useradd -r -s /usr/sbin/nologin lifeos 2>/dev/null || true
chown -R lifeos:lifeos "$APP_DIR"
cat > /etc/systemd/system/lifeos.service <<SVC
[Unit]
Description=LifeOS server
After=network-online.target postgresql.service
Wants=network-online.target
[Service]
Type=simple
User=lifeos
Group=lifeos
EnvironmentFile=$APP_DIR/lifeos.env
ExecStart=$APP_DIR/bin/lifeos migrate up
ExecStart=$APP_DIR/bin/lifeos serve
Restart=on-failure
RestartSec=5
[Install]
WantedBy=multi-user.target
SVC

# --- systemd: cloudflared quick tunnel -> :8080 ----------------------------------
cat > /etc/systemd/system/lifeos-tunnel.service <<'SVC'
[Unit]
Description=LifeOS cloudflared quick tunnel
After=network-online.target
Wants=network-online.target
[Service]
ExecStart=/usr/bin/cloudflared --no-autoupdate --url http://127.0.0.1:8080 tunnel
Restart=always
RestartSec=3
DynamicUser=yes
CacheDirectory=cloudflared
[Install]
WantedBy=multi-user.target
SVC

# --- nginx reverse proxy (cloudflared connects via :80 Host-header routing) ------
cat > /etc/nginx/sites-available/lifeos.conf <<'NGX'
server {
    listen 80 default_server;
    server_name _;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto https;
    }
}
NGX
ln -sf /etc/nginx/sites-available/lifeos.conf /etc/nginx/sites-enabled/lifeos.conf
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl reload nginx

# --- firewall: only ssh + http; app stays on loopback -----------------------------
# NOTE: this sandbox/dev-box is NOT the admin host — do not enable ufw here.
if command -v ufw >/dev/null 2>&1 && [ "${ALLOW_UFW:-0}" = "1" ]; then
  ufw allow OpenSSH >/dev/null 2>&1 || true
  ufw allow 80/tcp  >/dev/null 2>&1 || true
  ufw --force enable >/dev/null 2>&1 || true
else
  log "Skipping UFW (set ALLOW_UFW=1 to enable; Yandex Cloud security groups are preferred)"
fi

systemctl daemon-reload
systemctl enable --now lifeos-tunnel.service
systemctl enable --now lifeos.service

sleep 3
URL=$(journalctl -u lifeos-tunnel --no-pager -n 50 2>/dev/null | grep -oE 'https://[a-z0-9.-]+\.trycloudflare\.com' | tail -1 || true)
echo "[install] Tunnel URL: ${URL:-не появился в логах — journalctl -u lifeos-tunnel}"
echo "[install] Дальше: впишите токен и LIFEOS_MINIAPP_URL=$URL/app/ в $APP_DIR/lifeos.env && systemctl restart lifeos"
