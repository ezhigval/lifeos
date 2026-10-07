#!/usr/bin/env bash
# Deploy LifeOS to Fly.io and wire Telegram Mini App / webhook URLs.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v fly >/dev/null 2>&1 && ! command -v flyctl >/dev/null 2>&1; then
  echo "Install flyctl: https://fly.io/docs/hands-on/install-flyctl/" >&2
  exit 1
fi
FLY=$(command -v fly || command -v flyctl)

APP="${FLY_APP:-lifeos}"
echo "==> fly app: $APP"

if ! $FLY status -a "$APP" >/dev/null 2>&1; then
  echo "==> first launch (will prompt for org/region if needed)"
  $FLY launch --copy-config --config fly.toml --name "$APP" --no-deploy --yes || true
fi

echo "==> ensure Postgres (fly postgres)"
if ! $FLY postgres list 2>/dev/null | grep -q lifeos; then
  echo "Create Postgres (interactive once):"
  echo "  fly postgres create --name lifeos-db --region ams"
  echo "  fly postgres attach lifeos-db -a $APP"
  echo "Then re-run this script."
fi

echo "==> secrets from ../.env (TELEGRAM_BOT_TOKEN, JWT, …)"
if [[ -f "$ROOT/.env" ]]; then
  # shellcheck disable=SC1091
  set -a && source "$ROOT/.env" && set +a
fi
: "${TELEGRAM_BOT_TOKEN:?TELEGRAM_BOT_TOKEN required}"
: "${LIFEOS_JWT_SECRET:?LIFEOS_JWT_SECRET required}"
: "${LIFEOS_API_KEY:?LIFEOS_API_KEY required}"
: "${LIFEOS_TELEGRAM_WEBHOOK_SECRET:=$(openssl rand -hex 24)}"

$FLY secrets set -a "$APP" \
  TELEGRAM_BOT_TOKEN="$TELEGRAM_BOT_TOKEN" \
  LIFEOS_JWT_SECRET="$LIFEOS_JWT_SECRET" \
  LIFEOS_API_KEY="$LIFEOS_API_KEY" \
  LIFEOS_TELEGRAM_WEBHOOK_SECRET="$LIFEOS_TELEGRAM_WEBHOOK_SECRET" \
  LIFEOS_TELEGRAM_MODE=webhook \
  LIFEOS_SEED_TIMEZONE="${LIFEOS_SEED_TIMEZONE:-Europe/Moscow}"

echo "==> deploy"
$FLY deploy -a "$APP" --config fly.toml --remote-only

HOST="$($FLY info -a "$APP" --json 2>/dev/null | python3 -c 'import sys,json; print(json.load(sys.stdin).get("Hostname",""))' 2>/dev/null || true)"
if [[ -z "$HOST" ]]; then
  HOST="${APP}.fly.dev"
fi
BASE="https://${HOST}"
MINIAPP="${BASE}/app/"
WEBHOOK="${BASE}/webhook/telegram"

$FLY secrets set -a "$APP" \
  LIFEOS_MINIAPP_URL="$MINIAPP" \
  LIFEOS_TELEGRAM_WEBHOOK_URL="$WEBHOOK"

# redeploy once so process sees new Mini App URL in menu registration
$FLY deploy -a "$APP" --config fly.toml --remote-only

echo "==> wire Telegram"
export TELEGRAM_BOT_TOKEN LIFEOS_MINIAPP_URL="$MINIAPP" LIFEOS_TELEGRAM_WEBHOOK_URL="$WEBHOOK" LIFEOS_TELEGRAM_WEBHOOK_SECRET
"$ROOT/scripts/set-telegram-urls.sh"

echo
echo "Done."
echo "  Mini App: $MINIAPP"
echo "  Webhook:  $WEBHOOK"
echo "Open bot → синяя кнопка / Menu → Mini App"
