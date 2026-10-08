#!/usr/bin/env bash
# Install the named Cloudflare Tunnel unit on the VM.
#
# Preferred: create the token file yourself, then run this script.
#   sudo install -d -m 700 /opt/lifeos/secrets
#   sudo tee /opt/lifeos/secrets/tunnel.env >/dev/null <<'EOF'
#   TUNNEL_TOKEN=<paste>
#   EOF
#   sudo chmod 600 /opt/lifeos/secrets/tunnel.env
#   sudo bash deployments/cloudflared/install-named-tunnel.sh
#
# Alternative: sudo TUNNEL_TOKEN='<paste>' bash install-named-tunnel.sh
# The script writes the token to /opt/lifeos/secrets/tunnel.env and unsets it.
# It never echoes the token and refuses to write inside a git work tree.
set +x
set -euo pipefail

SECRETS=/opt/lifeos/secrets
ENV_FILE="$SECRETS/tunnel.env"
BIN=/opt/lifeos/bin
SRC="$(cd "$(dirname "$0")" && pwd)"
REPO_UNIT="$SRC/../systemd/lifeos-tunnel.service"

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

install_cloudflared() {
  if command -v cloudflared >/dev/null 2>&1 || [[ -x /usr/local/bin/cloudflared ]]; then
    return 0
  fi
  local arch deb tmp
  arch="$(uname -m)"
  case "$arch" in
    x86_64) deb=cloudflared-linux-amd64.deb ;;
    aarch64|arm64) deb=cloudflared-linux-arm64.deb ;;
    *) echo "unsupported arch $arch" >&2; exit 1 ;;
  esac
  tmp="$(mktemp)"
  echo "[tunnel] downloading cloudflared ($deb)"
  curl -fsSL -o "$tmp" "https://github.com/cloudflare/cloudflared/releases/latest/download/${deb}"
  dpkg -i "$tmp"
  rm -f "$tmp"
}

write_token_from_env() {
  [[ -n "${TUNNEL_TOKEN:-}" ]] || return 0
  install -d -m 700 "$SECRETS"
  refuse_git "$ENV_FILE"
  python3 - "$ENV_FILE" <<'PY'
import os, re, sys
path = sys.argv[1]
token = os.environ.get("TUNNEL_TOKEN", "").strip()
if not token or token.lower().startswith("paste") or token.lower() in {"change-me", "changeme"}:
    sys.stderr.write("TUNNEL_TOKEN is empty or a placeholder\n")
    sys.exit(1)
# Unquoted on purpose: systemd EnvironmentFile and bash source must agree.
# Cloudflare connector tokens are base64url JWTs and fit this set.
if not re.fullmatch(r"[A-Za-z0-9._~+/=-]{20,}", token):
    sys.stderr.write("TUNNEL_TOKEN must be one line of token characters, length >= 20\n")
    sys.exit(1)
open(path, "w").write(f"TUNNEL_TOKEN={token}\n")
os.chmod(path, 0o600)
PY
  unset TUNNEL_TOKEN
}

token_ready() {
  [[ -f "$ENV_FILE" ]] || return 1
  python3 - "$ENV_FILE" <<'PY'
import sys
token = ""
for line in open(sys.argv[1]):
    if line.startswith("TUNNEL_TOKEN="):
        token = line.split("=", 1)[1].strip()
if len(token) >= 2 and token[0] == token[-1] and token[0] in "'\"":
    token = token[1:-1]
if not token or any(ch.isspace() for ch in token) or len(token) < 20:
    sys.exit(1)
PY
}

install -d -m 755 "$BIN"
install -m 755 "$SRC/run-tunnel.sh" "$BIN/run-tunnel.sh"
install -m 644 "$REPO_UNIT" /etc/systemd/system/lifeos-tunnel.service
install_cloudflared
write_token_from_env
systemctl daemon-reload

if token_ready; then
  refuse_git "$ENV_FILE"
  chmod 600 "$ENV_FILE"
  systemctl enable --now lifeos-tunnel
  systemctl restart lifeos-tunnel
  sleep 2
  systemctl is-active lifeos-tunnel
  echo "[tunnel] named tunnel service is active. Token file: $ENV_FILE (not printed)."
  echo "[tunnel] public hostname service must be HTTP http://127.0.0.1:8080 (the Go app, not nginx)."
else
  systemctl disable --now lifeos-tunnel >/dev/null 2>&1 || true
  echo "[tunnel] unit installed, not started."
  echo "[tunnel] write $ENV_FILE (chmod 600) with TUNNEL_TOKEN=..., then: sudo systemctl enable --now lifeos-tunnel"
fi
