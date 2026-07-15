#!/usr/bin/env bash
# Register Telegram webhook + Mini App menu button from env, and optionally
# DM a fresh inline web_app button to LIFEOS_NOTIFY_CHAT_ID / SEED telegram id.
set -euo pipefail

: "${TELEGRAM_BOT_TOKEN:?TELEGRAM_BOT_TOKEN required}"
: "${LIFEOS_MINIAPP_URL:?LIFEOS_MINIAPP_URL required (…/app/)}"

API="https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}"

echo "==> setChatMenuButton → $LIFEOS_MINIAPP_URL"
curl -sS -X POST "$API/setChatMenuButton" \
  -H 'Content-Type: application/json' \
  -d "$(python3 - <<PY
import json, os
print(json.dumps({
  "menu_button": {
    "type": "web_app",
    "text": "Mini App",
    "web_app": {"url": os.environ["LIFEOS_MINIAPP_URL"]},
  }
}))
PY
)" | python3 -m json.tool

if [[ -n "${LIFEOS_TELEGRAM_WEBHOOK_URL:-}" ]]; then
  : "${LIFEOS_TELEGRAM_WEBHOOK_SECRET:?LIFEOS_TELEGRAM_WEBHOOK_SECRET required for webhook}"
  echo "==> setWebhook → $LIFEOS_TELEGRAM_WEBHOOK_URL"
  curl -sS -X POST "$API/setWebhook" \
    -H 'Content-Type: application/json' \
    -d "$(python3 - <<PY
import json, os
print(json.dumps({
  "url": os.environ["LIFEOS_TELEGRAM_WEBHOOK_URL"],
  "secret_token": os.environ["LIFEOS_TELEGRAM_WEBHOOK_SECRET"],
  "allowed_updates": ["message", "callback_query"],
  "drop_pending_updates": False,
}))
PY
)" | python3 -m json.tool
fi

CHAT="${LIFEOS_NOTIFY_CHAT_ID:-${LIFEOS_SEED_TELEGRAM_ID:-}}"
if [[ -n "$CHAT" && "$CHAT" != "0" ]]; then
  echo "==> send fresh Mini App button to chat $CHAT"
  export CHAT
  curl -sS -X POST "$API/sendMessage" \
    -H 'Content-Type: application/json' \
    -d "$(python3 - <<PY
import json, os
print(json.dumps({
  "chat_id": int(os.environ["CHAT"]),
  "text": "📱 <b>Mini App</b>\nОткрой синей кнопкой ниже (не из старой истории чата).",
  "parse_mode": "HTML",
  "reply_markup": {
    "inline_keyboard": [[{
      "text": "📱 Открыть Mini App",
      "web_app": {"url": os.environ["LIFEOS_MINIAPP_URL"]},
    }]]
  },
}, ensure_ascii=False))
PY
)" >/dev/null
fi

echo "==> getWebhookInfo"
curl -sS "$API/getWebhookInfo" | python3 -m json.tool
echo "OK"
