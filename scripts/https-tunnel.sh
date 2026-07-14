#!/usr/bin/env bash
# Starts a public HTTPS tunnel to local LifeOS (:8080) for Telegram Mini App.
# Requires cloudflared. Prints LIFEOS_MINIAPP_URL and optionally writes .env.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${1:-8080}"
LOG="${TMPDIR:-/tmp}/lifeos-cloudflared.log"

if ! command -v cloudflared >/dev/null 2>&1; then
  echo "cloudflared not found. Install: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/" >&2
  exit 1
fi

pkill -f "cloudflared tunnel --url" >/dev/null 2>&1 || true
: >"$LOG"
cloudflared tunnel --url "http://127.0.0.1:${PORT}" >"$LOG" 2>&1 &
echo $! >"${TMPDIR:-/tmp}/lifeos-cloudflared.pid"

echo "waiting for tunnel URL..."
URL=""
for _ in $(seq 1 40); do
  URL=$(grep -oE 'https://[a-zA-Z0-9.-]+\.trycloudflare\.com' "$LOG" | head -n1 || true)
  if [[ -n "$URL" ]]; then
    break
  fi
  sleep 0.5
done

if [[ -z "$URL" ]]; then
  echo "failed to obtain tunnel URL; see $LOG" >&2
  exit 1
fi

MINIAPP="${URL}/app/"
echo "HTTPS tunnel: $URL"
echo "Mini App URL: $MINIAPP"

ENV_FILE="$ROOT/.env"
if [[ -f "$ENV_FILE" ]]; then
  if grep -q '^LIFEOS_MINIAPP_URL=' "$ENV_FILE"; then
    sed -i "s|^LIFEOS_MINIAPP_URL=.*|LIFEOS_MINIAPP_URL=${MINIAPP}|" "$ENV_FILE"
  else
    printf '\nLIFEOS_MINIAPP_URL=%s\n' "$MINIAPP" >>"$ENV_FILE"
  fi
  echo "updated $ENV_FILE"
fi
