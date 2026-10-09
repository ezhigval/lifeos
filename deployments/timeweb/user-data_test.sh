#!/usr/bin/env bash
# Runs the cloud-init script without root, apt, or the network.
set -euo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
script="$root/deployments/timeweb/user-data.sh"
bash -n "$script"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

run_once() {
  LIFEOS_BOOTSTRAP_APPLY=0 \
    LIFEOS_BASE="$1" \
    LIFEOS_DOMAIN=local-ai-assist.ru \
    bash "$script" >"$2"
}

run_once "$tmp/base" "$tmp/out1"
# Secrets must not land in the log.
if grep -E '^[0-9a-f]{64}$' "$tmp/out1" >/dev/null; then
  echo "bootstrap printed a secret" >&2
  exit 1
fi
if grep -q '^LIFEOS_HTTP_PROXY=' "$tmp/base/.env"; then
  echo ".env must not set a telegram proxy" >&2
  exit 1
fi
grep -q '^LIFEOS_DOMAIN=local-ai-assist.ru$' "$tmp/base/.env"
grep -q '^TELEGRAM_BOT_TOKEN=$' "$tmp/base/.env"
grep -q '^CADDY_ACME_EMAIL=$' "$tmp/base/.env"
grep -q '^LIFEOS_TELEGRAM_MODE=webhook$' "$tmp/base/.env"

mode=$(stat -c '%a' "$tmp/base/.env")
if [[ "$mode" != "600" ]]; then
  echo ".env mode $mode" >&2
  exit 1
fi

pg=$(awk -F= '$1=="POSTGRES_PASSWORD" {print $2}' "$tmp/base/.env")
jwt=$(awk -F= '$1=="LIFEOS_JWT_SECRET" {print $2}' "$tmp/base/.env")
if [[ ${#pg} -ne 64 || ${#jwt} -ne 64 || "$pg" == "$jwt" ]]; then
  echo "expected two distinct 32-byte hex secrets" >&2
  exit 1
fi
if grep -q "$pg" "$tmp/out1" || grep -q "$jwt" "$tmp/out1"; then
  echo "secret leaked to stdout" >&2
  exit 1
fi

test -x "$tmp/base/bin/net-check.sh"
grep -q 'api.telegram.org' "$tmp/base/bin/net-check.sh"
grep -q 'TELEGRAM_BOT_TOKEN' "$tmp/base/NEXT.txt"

run_once "$tmp/base" "$tmp/out2"
pg2=$(awk -F= '$1=="POSTGRES_PASSWORD" {print $2}' "$tmp/base/.env")
if [[ "$pg" != "$pg2" ]]; then
  echo "second run replaced POSTGRES_PASSWORD" >&2
  exit 1
fi
grep -q 'already exists' "$tmp/out2"

echo "timeweb user-data tests ok"
