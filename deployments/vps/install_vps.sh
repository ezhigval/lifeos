#!/usr/bin/env bash
# LifeOS VPS installer (Yandex Cloud, Ubuntu 22.04/24.04). Run as root on a fresh VM:
#   sudo bash install.sh --source bundle    # source already rsync'ed to /opt/lifeos/src
# or with git:
#   REPO_URL=<git-url> sudo bash install.sh
# Secrets come from /root/lifeos-deploy.env (TG token, JWT secret, API key — copied from dev box).
set -euo pipefail

APP_DIR=/opt/lifeos
GO_VERSION=1.25.5
REPO_URL="${REPO_URL:-}"
SOURCE_MODE="${1:-}"   # "bundle" = source already in $APP_DIR/src; otherwise git clone/pull

log() { echo "[install] $*"; }
export DEBIAN_FRONTEND=noninteractive

apt-get update -y
apt-get install -y git curl ca-certificates gnupg postgresql nginx ufw rsync unzip gettext-base openssl

# --- PostgreSQL role/db -----------------------------------------------------
PG_PASSWORD="$(openssl rand -hex 16)"
sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='lifeos'" | grep -q 1 || \
  sudo -u postgres psql -c "CREATE ROLE lifeos LOGIN PASSWORD '${PG_PASSWORD}';"
sudo -u postgres psql -tAc "ALTER ROLE lifeos WITH PASSWORD '${PG_PASSWORD}';" >/dev/null
sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='lifeos'" | grep -q 1 || \
  sudo -u postgres createdb -O lifeos lifeos

# --- Go ---------------------------------------------------------------------
if ! /usr/local/go/bin/go version >/dev/null 2>&1; then
  log "Installing Go ${GO_VERSION}"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz
  rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tgz && rm /tmp/go.tgz
fi
export PATH="/usr/local/go/bin:$PATH"

# --- cloudflared --------------------------------------------------------------
if ! command -v cloudflared >/dev/null 2>&1; then
  log "Installing cloudflared"
  curl -fsSL https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb -o /tmp/cfd.deb
  apt-get install -y /tmp/cfd.deb && rm /tmp/cfd.deb
fi

# --- source -------------------------------------------------------------------
mkdir -p "$APP_DIR"
if [ "$SOURCE_MODE" = "--source" ] || [ "$SOURCE_MODE" = "bundle" ]; then
  log "Using pre-synced source in $APP_DIR/src"
elif [ -d "$APP_DIR/src/.git" ]; then
  [ -n "$REPO_URL" ] || REPO_URL="$(cat "$APP_DIR/repo_url" 2>/dev/null || true)"
  log "Pulling $REPO_URL"
  git -C "$APP_DIR/src" pull --ff-only
else
  [ -n "$REPO_URL" ] || { echo "REPO_URL required for first git-based install"; exit 1; }
  log "Cloning $REPO_URL"
  git clone "$REPO_URL" "$APP_DIR/src"
  echo "$REPO_URL" > "$APP_DIR/repo_url"
fi
cd "$APP_DIR/src"

# --- env ----------------------------------------------------------------------
DEPLOY_ENV=/root/lifeos-deploy.env
if [ ! -f "$DEPLOY_ENV" ]; then
  log "ERROR: $DEPLOY_ENV not found. Copy it from the dev box first:"
  log "  scp deployments/vps/lifeos-deploy.env.example root@VM:/root/lifeos-deploy.env"
  exit 1
fi
chmod 600 "$DEPLOY_ENV"
set -a; . "$DEPLOY_ENV"; set +a
export PG_PASSWORD
JWT_SECRET="${LIFEOS_JWT_SECRET:-$(openssl rand -hex 32)}"
API_KEY="${LIFEOS_API_KEY:-$(openssl rand -hex 16)}"
LEARNING_SALT="${LIFEOS_LEARNING_SALT:-$(openssl rand -hex 16)}"
export TG_BOT_TOKEN="$TELEGRAM_BOT_TOKEN" JWT_SECRET API_KEY LEARNING_SALT

envsubst < deployments/vps/lifeos.env.template > "$APP_DIR/lifeos.env"
chmod 600 "$APP_DIR/lifeos.env"
log "Wrote $APP_DIR/lifeos.env (miniapp URL placeholder — will be fixed after tunnel starts)"

# --- build --------------------------------------------------------------------
log "Building binary"
go build -o "$APP_DIR/bin/lifeos" ./cmd/lifeos

if command -v npm >/dev/null 2>&1 && [ -f web/miniapp/package.json ]; then
  log "Building miniapp (npm present)"
  (cd web/miniapp && npm ci --no-audit --no-fund && npm run build)
elif [ -d web/miniapp/dist ]; then
  log "MiniApp dist shipped with source bundle — skipping npm build"
fi

# --- migrations (run as app user via systemd ExecStartPre too) -----------------
log "Running migrations"
set -a; . "$APP_DIR/lifeos.env"; set +a
go run ./cmd/lifeos migrate up 2>/dev/null || {
  goose -dir migrations postgres "$LIFEOS_DATABASE_URL" up || true
}

# --- systemd: lifeos ------------------------------------------------------------
cat > /etc/systemd/system/lifeos.service <<'EOF'
[Unit]
Description=LifeOS server (bot + API + Mini App)
After=network-online.target postgresql.service
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=/opt/lifeos/lifeos.env
WorkingDirectory=/opt/lifeos/src
ExecStart=/opt/lifeos/bin/lifeos serve
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

# --- systemd: cloudflared quick-tunnel -----------------------------------------
TUNNEL_LOG=/var/log/lifeos-tunnel.log
cat > /etc/systemd/system/lifeos-tunnel.service <<EOF
[Unit]
Description=LifeOS cloudflared quick tunnel (80 -> 127.0.0.1:8080)
After=network-online.target

[Service]
Type=simple
ExecStartPre=/bin/bash -c 'cloudflared tunnel --no-autoupdate --url http://127.0.0.1:8080 2>&1 | tee ${TUNNEL_LOG}'
ExecStart=/bin/bash -c 'sleep infinity'
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
# simpler & robust: run cloudflared directly, parse URL in an ExecStart wrapper
cat > /opt/lifeos/tunnel.sh <<'EOF'
#!/usr/bin/env bash
# Runs cloudflared quick tunnel and publishes the resolved URL into lifeos.env + restarts app.
LOG=/var/log/lifeos-tunnel.log
cloudflared tunnel --no-autoupdate --url http://127.0.0.1:8080 2>"$LOG" &
CFD_PID=$!
URL=""
for i in $(seq 1 60); do
  URL=$(grep -oE 'https://[a-z0-9-]+\.trycloudflare\.com' "$LOG" | head -1 || true)
  [ -n "$URL" ] && break
  sleep 1
done
if [ -n "$URL" ]; then
  sed -i "s#^LIFEOS_MINIAPP_URL=.*#LIFEOS_MINIAPP_URL=${URL}/app/#" /opt/lifeos/lifeos.env
  systemctl restart lifeos
  echo "[tunnel] published $URL" >> "$LOG"
fi
wait $CFD_PID
EOF
chmod +x /opt/lifeos/tunnel.sh
cat > /etc/systemd/system/lifeos-tunnel.service <<'EOF'
[Unit]
Description=LifeOS cloudflared quick tunnel + URL publisher
After=network-online.target

[Service]
Type=simple
ExecStart=/opt/lifeos/tunnel.sh
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# --- nginx: public :80 -> :8080 (IP-based access as backup channel) -------------
cat > /etc/nginx/sites-available/lifeos <<'EOF'
server {
    listen 80 default_server;
    server_name _;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 120s;
    }
}
EOF
ln -sf /etc/nginx/sites-available/lifeos /etc/nginx/sites-enabled/lifeos
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl reload nginx || systemctl restart nginx

# --- UFW: SSH + HTTP only ---------------------------------------------------------
ufw allow 22/tcp
ufw allow 80/tcp
ufw --force enable

# --- services ---------------------------------------------------------------------
systemctl daemon-reload
systemctl enable --now lifeos-tunnel
sleep 12   # let tunnel publish URL
systemctl enable --now lifeos
sleep 3

log "Done."
URL=$(grep -oE 'https://[a-z0-9-]+\.trycloudflare\.com' "$TUNNEL_LOG" | head -1 || echo '<pending>')
echo "=================================================================="
echo " Tunnel URL : $URL"
echo " Mini App   : $URL/app/"
echo " Health     : $URL/health   (also http://$(curl -s ifconfig.me || echo VM-IP)/health)"
echo " Services   : systemctl status lifeos lifeos-tunnel"
echo " Logs       : journalctl -u lifeos -f ; tail -f $TUNNEL_LOG"
echo " NOTE       : after tunnel host changes -> send /start to the bot"
echo "=================================================================="
