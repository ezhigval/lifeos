#!/usr/bin/env bash
# Point the running VM at a stable public origin.
#
#   sudo LIFEOS_PUBLIC_ORIGIN=https://lifeos.example.com \
#        bash /opt/lifeos/repo/deployments/apply-public-origin.sh
#
# Writes only /opt/lifeos/.env (or /opt/lifeos/lifeos.env). Refuses paths inside
# a git work tree. Prints public URLs, never the webhook secret or bot token.
#
# Registers setWebhook and the menu button through the local proxy on
# 127.0.0.1:8081 so the call does not depend on direct egress to Telegram.
# Skip that with LIFEOS_SKIP_SETWEBHOOK=1.
set +x
set -euo pipefail

ORIGIN="${LIFEOS_PUBLIC_ORIGIN:-}"
ORIGIN="${ORIGIN%/}"
if [[ ! "$ORIGIN" =~ ^https://[A-Za-z0-9.-]+$ ]]; then
  echo "LIFEOS_PUBLIC_ORIGIN must look like https://lifeos.example.com" >&2
  exit 1
fi

if [[ -f /opt/lifeos/.env ]]; then
  ENV_FILE=/opt/lifeos/.env
elif [[ -f /opt/lifeos/lifeos.env ]]; then
  ENV_FILE=/opt/lifeos/lifeos.env
else
  echo "missing /opt/lifeos/.env" >&2
  exit 1
fi

real="$(realpath "$ENV_FILE")"
if git -C "$(dirname "$real")" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "refusing to edit $real: it is inside a git work tree" >&2
  exit 1
fi

export LIFEOS_ENV_FILE="$ENV_FILE"
export LIFEOS_PUBLIC_ORIGIN="$ORIGIN"
python3 <<'PY'
import os, secrets, sys
path = os.environ["LIFEOS_ENV_FILE"]
origin = os.environ["LIFEOS_PUBLIC_ORIGIN"].rstrip("/")
placeholders = {
    "",
    "change-me",
    "change-me-in-production",
    "change-me-long-random",
}
lines = open(path).read().splitlines()
vals = {}
for line in lines:
    if not line or line.startswith("#") or "=" not in line:
        continue
    k, v = line.split("=", 1)
    vals[k] = v

def setv(key, value):
    vals[key] = value

setv("LIFEOS_TELEGRAM_MODE", "webhook")
setv("LIFEOS_MINIAPP_URL", origin + "/app/")
setv("LIFEOS_TELEGRAM_WEBHOOK_URL", origin + "/webhook/telegram")
secret = vals.get("LIFEOS_TELEGRAM_WEBHOOK_SECRET", "").strip()
generated = False
if secret.lower() in placeholders:
    secret = secrets.token_hex(32)
    generated = True
    setv("LIFEOS_TELEGRAM_WEBHOOK_SECRET", secret)

seen = set()
out = []
for line in lines:
    if not line or line.startswith("#") or "=" not in line:
        out.append(line)
        continue
    k = line.split("=", 1)[0]
    if k in vals and k not in seen:
        out.append(f"{k}={vals[k]}")
        seen.add(k)
for k, v in vals.items():
    if k not in seen:
        out.append(f"{k}={v}")
        seen.add(k)
open(path, "w").write("\n".join(out) + "\n")
os.chmod(path, 0o600)
# Tell the shell whether a secret was generated, without printing it.
open("/tmp/lifeos-origin-generated", "w").write("1\n" if generated else "0\n")
PY
generated="$(cat /tmp/lifeos-origin-generated)"
rm -f /tmp/lifeos-origin-generated

echo "public origin: $ORIGIN"
echo "mini app:      $ORIGIN/app/"
echo "webhook:       $ORIGIN/webhook/telegram"
if [[ "$generated" == "1" ]]; then
  echo "webhook secret: generated and stored in $ENV_FILE"
else
  echo "webhook secret: kept existing value in $ENV_FILE"
fi

compose="/opt/lifeos/repo/deployments/docker-compose.yml"
if [[ -f "$compose" && -f /opt/lifeos/.env ]]; then
  args=(docker compose --env-file /opt/lifeos/.env -p lifeos -f "$compose")
  if [[ -f /opt/lifeos/docker-compose.override.yml ]]; then
    args+=(-f /opt/lifeos/docker-compose.override.yml)
  fi
  "${args[@]}" up -d --no-build --force-recreate app
elif systemctl cat lifeos.service >/dev/null 2>&1; then
  systemctl restart lifeos
fi

ok=0
for _ in 1 2 3 4 5 6 7 8 9 10 11 12; do
  if curl -fsS --max-time 3 http://127.0.0.1:8080/health >/dev/null; then
    ok=1
    break
  fi
  sleep 2
done
if [[ "$ok" != "1" ]]; then
  echo "local health did not become ready on 127.0.0.1:8080" >&2
  exit 1
fi
echo "local health: ok"

if [[ "${LIFEOS_SKIP_SETWEBHOOK:-}" == "1" ]]; then
  echo "setWebhook skipped"
else
  python3 - "$ENV_FILE" <<'PY'
import json, os, sys, urllib.request
path = sys.argv[1]
vals = {}
for line in open(path):
    line = line.strip()
    if not line or line.startswith("#") or "=" not in line:
        continue
    k, v = line.split("=", 1)
    vals[k] = v
token = vals.get("TELEGRAM_BOT_TOKEN") or vals.get("LIFEOS_TELEGRAM_BOT_TOKEN") or ""
webhook = vals.get("LIFEOS_TELEGRAM_WEBHOOK_URL", "")
secret = vals.get("LIFEOS_TELEGRAM_WEBHOOK_SECRET", "")
mini = vals.get("LIFEOS_MINIAPP_URL", "")
if not token or not webhook or not secret:
    sys.stderr.write("missing token, webhook url, or webhook secret in app env\n")
    sys.exit(1)
proxy = "http://127.0.0.1:8081"
opener = urllib.request.build_opener(urllib.request.ProxyHandler({"http": proxy}))

def post(method, payload):
    url = "http://api.telegram.org/bot" + token + "/" + method
    data = json.dumps(payload).encode()
    req = urllib.request.Request(url, data=data, method="POST")
    req.add_header("Content-Type", "application/json")
    try:
        with opener.open(req, timeout=30) as resp:
            body = json.loads(resp.read().decode())
    except Exception:
        sys.stderr.write(method + " via local proxy failed. Token not logged.\n")
        sys.exit(1)
    if not body.get("ok"):
        desc = body.get("description") or "not ok"
        sys.stderr.write(method + " rejected: " + str(desc) + "\n")
        sys.exit(1)

post("setWebhook", {
    "url": webhook,
    "secret_token": secret,
    "allowed_updates": ["message", "callback_query"],
    "drop_pending_updates": False,
})
print("setWebhook: ok")
if mini:
    post("setChatMenuButton", {
        "menu_button": {"type": "web_app", "text": "Mini App", "web_app": {"url": mini}},
    })
    print("menu button: ok")
info_url = "http://api.telegram.org/bot" + token + "/getWebhookInfo"
try:
    with opener.open(info_url, timeout=20) as resp:
        info = json.loads(resp.read().decode())
except Exception:
    sys.stderr.write("getWebhookInfo via local proxy failed. Token not logged.\n")
    sys.exit(1)
result = info.get("result") or {}
print(json.dumps({
    "webhook_url": result.get("url"),
    "pending": result.get("pending_update_count"),
    "last_error": result.get("last_error_message"),
}, ensure_ascii=False))
PY
fi

public_code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 15 "$ORIGIN/health" || echo fail)"
echo "public health: $public_code"
echo "done"
