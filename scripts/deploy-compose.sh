#!/usr/bin/env bash
# Deploy LifeOS on a VPS with docker compose + Caddy (Let's Encrypt).
# Usage: ./scripts/deploy-compose.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ROOT}/deployments/.env.prod"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing ${ENV_FILE}. Copy deployments/.env.prod.example and fill it in." >&2
  exit 1
fi

# shellcheck disable=SC1090
set -a
# shellcheck source=/dev/null
source "$ENV_FILE"
set +a

: "${LIFEOS_DOMAIN:?set LIFEOS_DOMAIN in .env.prod}"
: "${TELEGRAM_BOT_TOKEN:?set TELEGRAM_BOT_TOKEN in .env.prod}"
: "${LIFEOS_JWT_SECRET:?set LIFEOS_JWT_SECRET in .env.prod}"
: "${LIFEOS_TELEGRAM_WEBHOOK_SECRET:?set LIFEOS_TELEGRAM_WEBHOOK_SECRET in .env.prod}"
: "${POSTGRES_PASSWORD:?set POSTGRES_PASSWORD in .env.prod}"

BASE="https://${LIFEOS_DOMAIN}"
export LIFEOS_MINIAPP_URL="${LIFEOS_MINIAPP_URL:-${BASE}/app/}"
export LIFEOS_TELEGRAM_WEBHOOK_URL="${LIFEOS_TELEGRAM_WEBHOOK_URL:-${BASE}/webhook/telegram}"
export TELEGRAM_BOT_TOKEN LIFEOS_TELEGRAM_WEBHOOK_SECRET
export LIFEOS_NOTIFY_CHAT_ID="${LIFEOS_NOTIFY_CHAT_ID:-${LIFEOS_SEED_TELEGRAM_ID:-}}"

cd "${ROOT}/deployments"
docker compose -f docker-compose.prod.yml --env-file "$ENV_FILE" up -d --build

echo "Waiting for ${BASE}/health ..."
ok=0
for _ in $(seq 1 45); do
  if curl -fsS "${BASE}/health" >/dev/null 2>&1; then
    ok=1
    break
  fi
  sleep 2
done
if [[ "$ok" -ne 1 ]]; then
  echo "WARN: ${BASE}/health not ready yet — check DNS/ports 80+443 and compose logs." >&2
fi

"${ROOT}/scripts/set-telegram-urls.sh"

echo
echo "Deployed."
echo "  Mini App: $LIFEOS_MINIAPP_URL"
echo "  Webhook:  $LIFEOS_TELEGRAM_WEBHOOK_URL"
echo "Open Menu / the new blue button (not old chat history)."
