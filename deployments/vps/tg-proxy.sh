#!/usr/bin/env bash
# tg-proxy.sh — free Telegram API egress INSIDE the VM using a Cloudflare Worker as proxy.
#
# Why: Yandex Cloud blocks outbound to api.telegram.org, but does NOT block Cloudflare
# edge (workers.dev). We run a tiny local HTTP forward proxy on 127.0.0.1:8081 that
# rewrites api.telegram.org calls into a Worker URL; the Worker replays them against the
# real Telegram API from CF's anycast network. The app uses LIFEOS_HTTP_PROXY=http://127.0.0.1:8081.
# Everything is free and stays inside the VM (only external hop = Cloudflare itself).
#
# Prereq: deploy tg-proxy-worker.js as a Cloudflare Worker (dash.cloudflare.com > Workers >
# Create > paste code > Deploy, or `npx wrangler deploy` from deployments/vps/), then take
# its URL: https://tg-proxy.<your-subdomain>.workers.dev
#
# Usage (on the VM, as root):
#   LIFEOS_TG_PROXY_WORKER_URL=https://tg-proxy.<sub>.workers.dev ./tg-proxy.sh install|remove|test
set -euo pipefail

WORKER_URL="${LIFEOS_TG_PROXY_WORKER_URL:-}"
PROXY_PORT="${TG_PROXY_PORT:-8081}"
ENV_FILE="/opt/lifeos/lifeos.env"

if [[ -z "$WORKER_URL" ]]; then
  echo "usage: LIFEOS_TG_PROXY_WORKER_URL=https://tg-proxy.<sub>.workers.dev $0 install|remove|test" >&2
  exit 1
fi

install() {
  mkdir -p /opt/lifeos
  cp "$(dirname "$0")/tg-proxy.py" /opt/lifeos/tg-proxy.py
  chmod +x /opt/lifeos/tg-proxy.py

  cat > /etc/systemd/system/lifeos-tg-proxy.service <<EOF
[Unit]
Description=LifeOS Telegram API proxy via Cloudflare Worker
After=network-online.target

[Service]
ExecStart=/usr/bin/python3 /opt/lifeos/tg-proxy.py ${WORKER_URL} ${PROXY_PORT}
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
  systemctl enable --now lifeos-tg-proxy

  # wire into app env + restart app
  touch "$ENV_FILE"
  grep -q '^LIFEOS_HTTP_PROXY=' "$ENV_FILE" \
    && sed -i "s|^LIFEOS_HTTP_PROXY=.*|LIFEOS_HTTP_PROXY=http://127.0.0.1:${PROXY_PORT}|" "$ENV_FILE" \
    || echo "LIFEOS_HTTP_PROXY=http://127.0.0.1:${PROXY_PORT}" >> "$ENV_FILE"
  grep -q '^LIFEOS_TG_PROXY_WORKER_URL=' "$ENV_FILE" \
    && sed -i "s|^LIFEOS_TG_PROXY_WORKER_URL=.*|LIFEOS_TG_PROXY_WORKER_URL=${WORKER_URL}|" "$ENV_FILE" \
    || echo "LIFEOS_TG_PROXY_WORKER_URL=${WORKER_URL}" >> "$ENV_FILE"
  systemctl restart lifeos
  sleep 5
  curl -sf http://127.0.0.1:8080/health >/dev/null && echo "[tg-proxy] installed, app healthy"
}

test_proxy() {
  TOKEN=$(grep '^LIFEOS_TELEGRAM_BOT_TOKEN=' "$ENV_FILE" | cut -d= -f2-)
  [[ -z "$TOKEN" ]] && { echo "no token in $ENV_FILE"; exit 1; }
  echo "getMe via local proxy:"
  curl -sS --proxy "http://127.0.0.1:${PROXY_PORT}" \
    -X POST "http://api.telegram.org/bot${TOKEN}/getMe" | head -c 400; echo
}

remove() {
  systemctl disable --now lifeos-tg-proxy || true
  rm -f /etc/systemd/system/lifeos-tg-proxy.service /opt/lifeos/tg-proxy.py
  sed -i '/^LIFEOS_HTTP_PROXY=/d' "$ENV_FILE"
  systemctl restart lifeos
}

case "${1:-install}" in
  install) install ;;
  remove)  remove ;;
  test)    test_proxy ;;
  *) echo "usage: $0 install|remove|test" >&2; exit 1 ;;
esac
