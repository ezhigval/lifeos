#!/usr/bin/env bash
# Install the local Telegram egress proxy on the VM.
#
#   sudo LIFEOS_TG_PROXY_WORKER_URL=https://<name>.<account>.workers.dev \
#        LIFEOS_TG_PROXY_SECRET='<same value as wrangler secret PROXY_SECRET>' \
#        ./tg-proxy.sh install
#
# Secret is optional. When set, it is written only to /opt/lifeos/secrets/tg-proxy.env
# (mode 600) and never echoed. The app .env receives LIFEOS_HTTP_PROXY only.
#
#   sudo ./tg-proxy.sh test
#   sudo ./tg-proxy.sh remove
set +x
set -euo pipefail

SECRETS_DIR=/opt/lifeos/secrets
PROXY_ENV="$SECRETS_DIR/tg-proxy.env"
BIN_DIR=/opt/lifeos/bin
PORT="${TG_PROXY_PORT:-8081}"
SRC="$(cd "$(dirname "$0")" && pwd)"

usage() {
  echo "usage: LIFEOS_TG_PROXY_WORKER_URL=https://<worker> $0 install|remove|test" >&2
  exit 1
}

refuse_git() {
  local f="$1"
  mkdir -p "$(dirname "$f")"
  local dir
  dir="$(dirname "$(realpath -m "$f")")"
  if git -C "$dir" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "refusing to write $f inside a git work tree" >&2
    exit 1
  fi
}

app_env_file() {
  if [[ -f /opt/lifeos/.env ]]; then
    echo /opt/lifeos/.env
  elif [[ -f /opt/lifeos/lifeos.env ]]; then
    echo /opt/lifeos/lifeos.env
  else
    echo /opt/lifeos/.env
  fi
}

upsert_env() {
  local file="$1" key="$2" value="$3"
  refuse_git "$file"
  python3 - "$file" "$key" "$value" <<'PY'
import os, sys
path, key, value = sys.argv[1], sys.argv[2], sys.argv[3]
lines = open(path).read().splitlines() if os.path.exists(path) else []
out, found = [], False
for line in lines:
    if line.startswith(key + "="):
        out.append(f"{key}={value}")
        found = True
    else:
        out.append(line)
if not found:
    out.append(f"{key}={value}")
open(path, "w").write("\n".join(out) + "\n")
os.chmod(path, 0o600)
PY
}

delete_env() {
  local file="$1" key="$2"
  [[ -f "$file" ]] || return 0
  python3 - "$file" "$key" <<'PY'
import os, sys
path, key = sys.argv[1], sys.argv[2]
lines = [ln for ln in open(path).read().splitlines() if not ln.startswith(key + "=")]
open(path, "w").write("\n".join(lines) + ("\n" if lines else ""))
os.chmod(path, 0o600)
PY
}

docker_up() {
  local compose="/opt/lifeos/repo/deployments/docker-compose.yml"
  local envf="/opt/lifeos/.env"
  [[ -f "$compose" && -f "$envf" ]] || return 1
  local -a cmd=(docker compose --env-file "$envf" -p lifeos -f "$compose")
  if [[ -f /opt/lifeos/docker-compose.override.yml ]]; then
    cmd+=(-f /opt/lifeos/docker-compose.override.yml)
  fi
  "${cmd[@]}" "$@"
}

ensure_extra_hosts() {
  [[ -f /opt/lifeos/repo/deployments/docker-compose.yml ]] || return 0
  docker network inspect lifeos_default >/dev/null 2>&1 || return 0
  local override=/opt/lifeos/docker-compose.override.yml
  refuse_git "$override"
  python3 - "$override" <<'PY'
import os, sys
path = sys.argv[1]
text = open(path).read() if os.path.exists(path) else "services:\n  app:\n"
if "host.docker.internal" in text:
    raise SystemExit(0)
lines = text.splitlines()
out, inserted = [], False
for line in lines:
    out.append(line)
    if not inserted and line == "  app:":
        out.append("    extra_hosts:")
        out.append('      - "host.docker.internal:host-gateway"')
        inserted = True
if not inserted:
    if not any(ln.strip() == "services:" for ln in lines):
        out = ["services:", "  app:"]
    out.append("    extra_hosts:")
    out.append('      - "host.docker.internal:host-gateway"')
open(path, "w").write("\n".join(out).rstrip() + "\n")
PY
  echo "[tg-proxy] added host.docker.internal to $override"
}

restart_app() {
  if docker_up up -d --no-build --force-recreate app; then
    return 0
  fi
  if systemctl cat lifeos.service >/dev/null 2>&1; then
    systemctl restart lifeos
  fi
}

write_proxy_env() {
  local worker="$1"
  install -d -m 700 "$SECRETS_DIR"
  refuse_git "$PROXY_ENV"
  # Secret stays in the environment of this process. It is not passed as argv.
  python3 - "$PROXY_ENV" "$worker" "$PORT" <<'PY'
import os, sys
path, worker, port = sys.argv[1:4]
secret = os.environ.get("LIFEOS_TG_PROXY_SECRET", "").strip()

def q(value: str) -> str:
    return "'" + value.replace("'", "'\\''") + "'"

lines = [
    f"LIFEOS_TG_PROXY_WORKER_URL={q(worker)}",
    f"TG_PROXY_PORT={port}",
    "TG_PROXY_BIND=127.0.0.1",
]
if secret:
    lines.append(f"LIFEOS_TG_PROXY_SECRET={q(secret)}")
open(path, "w").write("\n".join(lines) + "\n")
os.chmod(path, 0o600)
PY
}

install_proxy() {
  local worker="${LIFEOS_TG_PROXY_WORKER_URL:-}"
  [[ -n "$worker" ]] || usage
  case "$worker" in
    https://*) ;;
    *) echo "worker URL must start with https://" >&2; exit 1 ;;
  esac

  install -d -m 755 "$BIN_DIR"
  install -m 755 "$SRC/tg-proxy.py" /opt/lifeos/tg-proxy.py
  install -m 755 "$SRC/tg-proxy-run.sh" "$BIN_DIR/tg-proxy-run.sh"
  install -m 644 "$SRC/../systemd/lifeos-tg-proxy.service" /etc/systemd/system/lifeos-tg-proxy.service
  write_proxy_env "$worker"
  # Drop the secret from this process environment before any further logs.
  unset LIFEOS_TG_PROXY_SECRET || true

  local envf proxy_url
  envf="$(app_env_file)"
  touch "$envf"
  chmod 600 "$envf"
  if docker network inspect lifeos_default >/dev/null 2>&1; then
    ensure_extra_hosts
    proxy_url="http://host.docker.internal:${PORT}"
  else
    proxy_url="http://127.0.0.1:${PORT}"
  fi
  upsert_env "$envf" LIFEOS_HTTP_PROXY "$proxy_url"
  upsert_env "$envf" LIFEOS_TG_PROXY_WORKER_URL "$worker"

  systemctl daemon-reload
  systemctl enable --now lifeos-tg-proxy
  systemctl restart lifeos-tg-proxy
  sleep 1
  restart_app || echo "[tg-proxy] app restart skipped (compose/service not found)"
  echo "[tg-proxy] installed. Proxy for the app: ${proxy_url}"
  echo "[tg-proxy] run: $0 test"
}

test_proxy() {
  local envf
  envf="$(app_env_file)"
  [[ -f "$envf" ]] || { echo "no app env at $envf" >&2; exit 1; }
  python3 - "$envf" "$PORT" <<'PY'
import json, os, sys, urllib.request
path, port = sys.argv[1], sys.argv[2]
vals = {}
for line in open(path):
    line = line.strip()
    if not line or line.startswith("#") or "=" not in line:
        continue
    k, v = line.split("=", 1)
    vals[k] = v
token = vals.get("TELEGRAM_BOT_TOKEN") or vals.get("LIFEOS_TELEGRAM_BOT_TOKEN") or ""
if not token:
    sys.stderr.write("no TELEGRAM_BOT_TOKEN in app env\n")
    sys.exit(1)
proxy = f"http://127.0.0.1:{port}"
url = "http://api.telegram.org/bot" + token + "/getMe"
opener = urllib.request.build_opener(urllib.request.ProxyHandler({"http": proxy}))
try:
    with opener.open(url, data=b"{}", timeout=20) as resp:
        body = resp.read()
except Exception:
    sys.stderr.write("getMe via local proxy failed (worker unreachable or secret mismatch). Token not logged.\n")
    sys.exit(1)
try:
    data = json.loads(body)
except json.JSONDecodeError:
    sys.stderr.write("getMe returned non-json\n")
    sys.exit(1)
result = data.get("result") or {}
print(json.dumps({
    "ok": data.get("ok"),
    "id": result.get("id"),
    "username": result.get("username"),
    "description": data.get("description"),
}, ensure_ascii=False))
PY
}

remove_proxy() {
  systemctl disable --now lifeos-tg-proxy >/dev/null 2>&1 || true
  rm -f /etc/systemd/system/lifeos-tg-proxy.service
  systemctl daemon-reload
  local envf
  envf="$(app_env_file)"
  if [[ -f "$envf" ]]; then
    delete_env "$envf" LIFEOS_HTTP_PROXY
    delete_env "$envf" LIFEOS_TG_PROXY_WORKER_URL
    restart_app || true
  fi
  echo "[tg-proxy] removed. Secrets file left at $PROXY_ENV (delete it yourself if you want)."
}

cmd="${1:-install}"
case "$cmd" in
  install) install_proxy ;;
  remove)  remove_proxy ;;
  test)    test_proxy ;;
  *) usage ;;
esac
