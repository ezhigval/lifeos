#!/usr/bin/env bash
# Named Cloudflare Tunnel. Token comes from the environment (systemd
# EnvironmentFile or /opt/lifeos/secrets/tunnel.env). It is not passed as an
# argv flag, so `ps` does not show it. cloudflared 2022.3+ reads TUNNEL_TOKEN.
set +x
set -euo pipefail

ENV_FILE="${LIFEOS_TUNNEL_ENV:-/opt/lifeos/secrets/tunnel.env}"
if [[ -z "${TUNNEL_TOKEN:-}" && -f "$ENV_FILE" ]]; then
  real=$(realpath "$ENV_FILE")
  if git -C "$(dirname "$real")" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "refusing to read tunnel token from a git work tree ($real)" >&2
    exit 1
  fi
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

if [[ -z "${TUNNEL_TOKEN:-}" ]]; then
  echo "TUNNEL_TOKEN is empty. Write it to $ENV_FILE (chmod 600), not to git." >&2
  exit 1
fi

CF="${CLOUDFLARED_BIN:-}"
if [[ -z "$CF" ]]; then
  CF=$(command -v cloudflared || true)
fi
if [[ -z "$CF" && -x /usr/local/bin/cloudflared ]]; then
  CF=/usr/local/bin/cloudflared
fi
if [[ -z "$CF" ]]; then
  echo "cloudflared is not installed" >&2
  exit 1
fi

# http2 over TCP 7844. QUIC (UDP 7844) is not reliable from this network,
# and the VM has no routable IPv6.
exec "$CF" --no-autoupdate --protocol http2 --edge-ip-version 4 tunnel run
