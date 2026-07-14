#!/usr/bin/env bash
# Verify LifeOS is reachable and initData → telegram_id → JWT path works.
# Usage:
#   ./scripts/verify-webapp-auth.sh              # health + /app + unit/e2e tests
#   ./scripts/verify-webapp-auth.sh http://127.0.0.1:8080
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASE="${1:-http://127.0.0.1:8080}"
BASE="${BASE%/}"

red() { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }
info() { printf '→ %s\n' "$*"; }

fail=0

info "1) Origin health ($BASE/health)"
if code=$(curl -sS -o /tmp/lifeos-health.json -w '%{http_code}' --connect-timeout 3 "$BASE/health" 2>/dev/null); then
  if [[ "$code" == "200" ]]; then
    green "OK /health → $code"
  else
    red "FAIL /health → HTTP $code (Cloudflare 1033 often means this origin is down)"
    fail=1
  fi
else
  red "FAIL cannot connect to $BASE — start app first (make docker-up / make dev)"
  red "Cloudflare Error 1033 = tunnel hostname up, but nothing on :8080"
  fail=1
fi

info "2) Mini App static ($BASE/app/)"
code=$(curl -sS -o /dev/null -w '%{http_code}' --connect-timeout 3 "$BASE/app/" 2>/dev/null || true; code=${code:-000})
if [[ "$code" == "200" ]]; then
  green "OK /app/ → $code"
else
  red "FAIL /app/ → HTTP $code (build miniapp + LIFEOS_STATIC_DIR)"
  fail=1
fi

info "3) Unit: initData HMAC + telegram user id parsing"
(
  cd "$ROOT"
  go test ./internal/platform/auth/ -count=1 -run 'ValidateWebAppInitData'
) || fail=1

info "4) HTTP: auth/telegram-webapp returns telegram_id from signed user.id"
(
  cd "$ROOT"
  go test ./internal/transport/http/api/ -count=1 -run 'AuthTelegramWebAppIssuesToken'
) || fail=1

info "5) Telegram keyboard: Mini App URL rotation forces reinstall"
(
  cd "$ROOT"
  go test ./internal/transport/telegram/ -count=1 -run 'ReplyKeyboard'
) || fail=1

if [[ -f "$ROOT/.env" ]]; then
  # shellcheck disable=SC1091
  set -a && source "$ROOT/.env" && set +a
fi

if [[ "${TELEGRAM_BOT_TOKEN:-}" != "" && "${LIFEOS_JWT_SECRET:-}" != "" && "$fail" -eq 0 && "$code" == "200" ]]; then
  info "6) Live POST /api/v1/auth/telegram-webapp against $BASE"
  LIVE_OUT=$(
    TELEGRAM_BOT_TOKEN="$TELEGRAM_BOT_TOKEN" BASE_URL="$BASE" go run "$ROOT/scripts/verify_webapp_auth_live.go" 2>&1
  ) || {
    red "$LIVE_OUT"
    fail=1
  }
  if [[ $fail -eq 0 ]]; then
    green "$LIVE_OUT"
  fi
else
  info "6) Skip live POST (need running server + TELEGRAM_BOT_TOKEN + LIFEOS_JWT_SECRET in .env)"
fi

echo
if [[ $fail -eq 0 ]]; then
  green "All webapp auth checks passed."
  echo "If Telegram still shows Cloudflare 1033:"
  echo "  1) make stack-up   # or: make docker-up && make tunnel && compose recreate app"
  echo "  2) In bot: /start  # refreshes reply keyboard with new LIFEOS_MINIAPP_URL"
  echo "  3) Open Mini App via the new «📱 Mini App» button (not an old chat link)"
  exit 0
fi
red "Some checks failed."
exit 1
