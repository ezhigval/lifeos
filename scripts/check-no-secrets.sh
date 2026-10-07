#!/usr/bin/env bash
# Fail if tracked files contain private keys, a filled tunnel token, or a bot token.
# The script itself is excluded so its own patterns are not matches.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

fail=0
scan() {
  local label="$1"
  shift
  local out
  if out=$(git grep -n -I -E "$@" -- . ':!scripts/check-no-secrets.sh' ':!web/miniapp/package-lock.json' 2>/dev/null); then
    echo "$label"
    echo "$out"
    fail=1
  fi
}

scan "private key material:" 'BEGIN (RSA |OPENSSH |EC |PGP )?PRIVATE KEY'
scan "filled TUNNEL_TOKEN (must stay empty in git):" '^TUNNEL_TOKEN=.+'
scan "filled proxy secret:" '^LIFEOS_TG_PROXY_SECRET=.+'
scan "possible Telegram bot token:" '[0-9]{8,12}:[A-Za-z0-9_-]{35}'

if [[ "$fail" -ne 0 ]]; then
  echo "refusing: secret-looking content is tracked. Keep tokens in /opt/lifeos/secrets and /opt/lifeos/.env on the VM." >&2
  exit 1
fi
echo "no tracked secrets matched"
