#!/usr/bin/env bash
# Starts tg-proxy.py with secrets from /opt/lifeos/secrets/tg-proxy.env.
# When the lifeos docker network exists, also binds the bridge gateway so the
# app container can reach the proxy via host.docker.internal. The public NIC
# is not bound. Never prints the env file.
set +x
set -euo pipefail

ENV_FILE="${LIFEOS_TG_PROXY_ENV:-/opt/lifeos/secrets/tg-proxy.env}"
PY="${LIFEOS_TG_PROXY_PY:-/opt/lifeos/tg-proxy.py}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "missing $ENV_FILE" >&2
  exit 1
fi
if [[ ! -f "$PY" ]]; then
  echo "missing $PY" >&2
  exit 1
fi

real=$(realpath "$ENV_FILE")
if git -C "$(dirname "$real")" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "refusing to read proxy secrets from a git work tree ($real)" >&2
  exit 1
fi

# host.docker.internal is docker0 (host-gateway), not the compose-network
# gateway. Bind every docker bridge, plus 127.0.0.1. Never the public NIC.
bind=$(python3 - <<'PY'
import os, subprocess
addrs = ["127.0.0.1"]
seen = set(addrs)
try:
    out = subprocess.check_output(["ip", "-4", "-o", "addr", "show"], text=True)
except Exception:
    out = ""
for line in out.splitlines():
    parts = line.split()
    if len(parts) < 4:
        continue
    iface, cidr = parts[1], parts[3]
    if iface != "docker0" and not iface.startswith("br-"):
        continue
    ip = cidr.split("/", 1)[0]
    if ip not in seen:
        addrs.append(ip)
        seen.add(ip)
if os.path.exists("/var/run/docker.sock") or os.path.exists("/run/docker.sock"):
    try:
        gw = subprocess.check_output(
            ["docker", "network", "inspect", "lifeos_default", "-f", "{{(index .IPAM.Config 0).Gateway}}"],
            text=True,
            stderr=subprocess.DEVNULL,
        ).strip()
    except Exception:
        gw = ""
    if gw and gw not in seen:
        addrs.append(gw)
print(",".join(addrs))
PY
)

python3 - "$ENV_FILE" "$bind" <<'PY'
import os, sys
path, bind = sys.argv[1], sys.argv[2]
lines = open(path).read().splitlines()
out, found = [], False
for line in lines:
    if line.startswith("TG_PROXY_BIND="):
        out.append(f"TG_PROXY_BIND={bind}")
        found = True
    else:
        out.append(line)
if not found:
    out.append(f"TG_PROXY_BIND={bind}")
text = "\n".join(out) + "\n"
open(path, "w").write(text)
os.chmod(path, 0o600)
PY

set -a
# shellcheck disable=SC1090
. "$ENV_FILE"
set +a

if [[ -z "${LIFEOS_TG_PROXY_WORKER_URL:-}" ]]; then
  echo "LIFEOS_TG_PROXY_WORKER_URL is empty in $ENV_FILE" >&2
  exit 1
fi

echo "tg-proxy bind ${TG_PROXY_BIND:-127.0.0.1}:${TG_PROXY_PORT:-8081}"
exec /usr/bin/python3 "$PY"
