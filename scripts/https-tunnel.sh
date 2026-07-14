#!/usr/bin/env bash
# Starts a public HTTPS tunnel to local LifeOS (:8080) for Telegram Mini App.
# Requires cloudflared. Prints LIFEOS_MINIAPP_URL and optionally writes .env.
#
# Cloudflare Error 1033 = public hostname exists but origin :8080 is unreachable.
# Always ensure `curl -sS http://127.0.0.1:8080/health` works before / after tunnel.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${1:-8080}"
LOG="${TMPDIR:-/tmp}/lifeos-cloudflared.log"
ORIGIN="http://127.0.0.1:${PORT}"

if ! command -v cloudflared >/dev/null 2>&1; then
  echo "cloudflared not found. Install: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/" >&2
  exit 1
fi

echo "checking origin $ORIGIN/health ..."
ORIGIN_OK=0
for _ in $(seq 1 30); do
  if curl -sS -o /dev/null -w '' --connect-timeout 1 "$ORIGIN/health" 2>/dev/null; then
    code=$(curl -sS -o /dev/null -w '%{http_code}' --connect-timeout 1 "$ORIGIN/health" || true)
    if [[ "$code" == "200" ]]; then
      ORIGIN_OK=1
      break
    fi
  fi
  sleep 0.5
done
if [[ "$ORIGIN_OK" -ne 1 ]]; then
  echo "origin $ORIGIN is not healthy." >&2
  echo "Start the app first: make docker-up   OR   make dev" >&2
  echo "Cloudflare Error 1033 means the tunnel cannot reach this origin." >&2
  exit 1
fi
echo "origin OK"

pkill -f "cloudflared tunnel --url" >/dev/null 2>&1 || true
: >"$LOG"
cloudflared tunnel --url "$ORIGIN" >"$LOG" 2>&1 &
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

echo
echo "Next:"
echo "  1) Recreate app so it loads LIFEOS_MINIAPP_URL (make stack-up does this)"
echo "  2) In Telegram send /start — refreshes reply keyboard with the NEW url"
echo "  3) Open Mini App only via the new button (old links → Error 1033)"
echo "  4) ./scripts/verify-webapp-auth.sh $ORIGIN"
