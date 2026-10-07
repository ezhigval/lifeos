#!/usr/bin/env bash
# Non-interactive Fly deploy for GitHub Actions (and local CI use).
# Requires: flyctl, FLY_API_TOKEN, TELEGRAM_BOT_TOKEN, LIFEOS_JWT_SECRET,
# LIFEOS_API_KEY, LIFEOS_TELEGRAM_WEBHOOK_SECRET.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FLY=$(command -v flyctl || command -v fly || true)
if [[ -z "$FLY" ]]; then
  echo "flyctl not found" >&2
  exit 1
fi

APP="${FLY_APP:-lifeos}"
WIRE_TELEGRAM="${WIRE_TELEGRAM:-true}"

: "${FLY_API_TOKEN:?FLY_API_TOKEN required}"
: "${TELEGRAM_BOT_TOKEN:?TELEGRAM_BOT_TOKEN required}"
: "${LIFEOS_JWT_SECRET:?LIFEOS_JWT_SECRET required}"
: "${LIFEOS_API_KEY:?LIFEOS_API_KEY required}"
: "${LIFEOS_TELEGRAM_WEBHOOK_SECRET:?LIFEOS_TELEGRAM_WEBHOOK_SECRET required}"

echo "==> fly app: $APP"

if ! $FLY status -a "$APP" >/dev/null 2>&1; then
  echo "Fly app '$APP' not found. Create once:" >&2
  echo "  fly apps create $APP" >&2
  echo "  fly postgres create --name ${APP}-db --region ams" >&2
  echo "  fly postgres attach ${APP}-db -a $APP" >&2
  exit 1
fi

echo "==> set app secrets"
$FLY secrets set -a "$APP" \
  TELEGRAM_BOT_TOKEN="$TELEGRAM_BOT_TOKEN" \
  LIFEOS_JWT_SECRET="$LIFEOS_JWT_SECRET" \
  LIFEOS_API_KEY="$LIFEOS_API_KEY" \
  LIFEOS_TELEGRAM_WEBHOOK_SECRET="$LIFEOS_TELEGRAM_WEBHOOK_SECRET" \
  LIFEOS_TELEGRAM_MODE=webhook \
  LIFEOS_SEED_TIMEZONE="${LIFEOS_SEED_TIMEZONE:-Europe/Moscow}" \
  LIFEOS_SEED_TELEGRAM_ID="${LIFEOS_SEED_TELEGRAM_ID:-0}" \
  LIFEOS_LLM_ENABLED="${LIFEOS_LLM_ENABLED:-false}" \
  LIFEOS_LLM_AGENT_ENABLED="${LIFEOS_LLM_AGENT_ENABLED:-false}"

HOST="$($FLY info -a "$APP" --json 2>/dev/null | python3 -c 'import sys,json; print(json.load(sys.stdin).get("Hostname") or "")' 2>/dev/null || true)"
if [[ -z "$HOST" ]]; then
  HOST="${APP}.fly.dev"
fi
BASE="https://${HOST}"
MINIAPP="${BASE}/app/"
WEBHOOK="${BASE}/webhook/telegram"

$FLY secrets set -a "$APP" \
  LIFEOS_MINIAPP_URL="$MINIAPP" \
  LIFEOS_TELEGRAM_WEBHOOK_URL="$WEBHOOK"

echo "==> deploy"
$FLY deploy -a "$APP" --config deployments/fly.toml --dockerfile deployments/Dockerfile --remote-only

echo "==> health"
ok=0
for _ in $(seq 1 30); do
  if curl -fsS "${BASE}/health" >/dev/null 2>&1; then
    ok=1
    break
  fi
  sleep 2
done
if [[ "$ok" -ne 1 ]]; then
  echo "WARN: ${BASE}/health not ready yet" >&2
else
  echo "OK ${BASE}/health"
fi

if [[ "$WIRE_TELEGRAM" == "true" || "$WIRE_TELEGRAM" == "1" ]]; then
  echo "==> wire Telegram webhook + menu"
  export TELEGRAM_BOT_TOKEN LIFEOS_MINIAPP_URL="$MINIAPP" \
    LIFEOS_TELEGRAM_WEBHOOK_URL="$WEBHOOK" LIFEOS_TELEGRAM_WEBHOOK_SECRET \
    LIFEOS_NOTIFY_CHAT_ID="${LIFEOS_NOTIFY_CHAT_ID:-${LIFEOS_SEED_TELEGRAM_ID:-}}"
  "$ROOT/scripts/set-telegram-urls.sh"
fi

echo
echo "Deployed."
echo "  Mini App: $MINIAPP"
echo "  Webhook:  $WEBHOOK"
