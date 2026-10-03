#!/usr/bin/env bash
# tg-proxy.sh — free Telegram API egress INSIDE the VM using Cloudflare Workers as an HTTP proxy.
#
# Why: Yandex Cloud blocks outbound to api.telegram.org (all DC IPs). But it does NOT block
# https://cloudflare.com / *.workers.dev (Cloudflare edge). So we run a tiny local forward proxy
# on 127.0.0.1:8081 that rewrites api.telegram.org requests into a Worker URL, and the app uses
# LIFEOS_HTTP_PROXY=http://127.0.0.1:8081. The Worker fetches the real Telegram API from CF's
# anycast network (unblocked) and returns the response. Everything stays free and inside the VM
# (the only external hop is Cloudflare itself, which YC allows).
#
# Components:
#   1) Worker code (tg-proxy-worker.js) -> deploy via dash.cloudflare.com > Workers > Create > Paste
#      or `npx wrangler deploy` from anywhere with CF auth. Note the worker URL:
#      https://tg-proxy.<your-subdomain>.workers.dev
#   2) Local proxy (this file, systemd unit lifeos-tg-proxy) on 127.0.0.1:8081
#   3) env: LIFEOS_HTTP_PROXY + LIFEOS_TG_PROXY_WORKER_URL in /opt/lifeos/lifeos.env
set -euo pipefail

WORKER_URL="${LIFEOS_TG_PROXY_WORKER_URL:-}"
PROXY_PORT="${TG_PROXY_PORT:-8081}"

if [[ -z "$WORKER_URL" ]]; then
  echo "usage: LIFEOS_TG_PROXY_WORKER_URL=https://tg-proxy.<sub>.workers.dev $0 install|remove" >&2
  exit 1
fi

case "${1:-install}" in
install)
  mkdir -p /opt/lifeos
  cat > /opt/lifeos/tg-proxy.py <<PYEOF
#!/usr/bin/env python3
"""Minimal CONNECT-less HTTP forward proxy that tunnels api.telegram.org via a Cloudflare Worker."""
import http.server, socketserver, urllib.request, json, sys, re

WORKER = sys.argv[1].rstrip("/")
PORT = int(sys.argv[2]) if len(sys.argv) > 2 else 8081

class H(http.server.BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def _target(self):
        # absolute-form URI from our client: http://api.telegram.org/botXXX/method
        u = self.path
        m = re.match(r"^http://api\.telegram\.org(/.*)$", u)
        return m.group(1) if m else None

    def do_POST(self):
        t = self._target()
        if not t:
            self.send_error(404); return
        n = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(n)
        req = urllib.request.Request(WORKER + "/fetch?path=" + t, data=body, method="POST",
                                     headers={"Content-Type": "application/json"})
        try:
            with urllib.request.urlopen(req, timeout=30) as r:
                data, code = r.read(), r.status
        except Exception as e:
            data, code = json.dumps({"ok": False, "error": str(e)}).encode(), 502
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        t = self._target()
        if not t:
            self.send_error(404); return
        req = urllib.request.Request(WORKER + "/fetch?path=" + t.lstrip("/"), method="GET")
        try:
            with urllib.request.urlopen(req, timeout=30) as r:
                data, code = r.read(), r.status
        except Exception as e:
            data, code = json.dumps({"ok": False, "error": str(e)}).encode(), 502
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def log_message(self, *a): pass

socketserver.ThreadingTCPServer.allow_reuse_address = True
with socketserver.ThreadingTCPServer(("127.0.0.1", PORT), H) as srv:
    srv.serve_forever()
PYEOF
  chmod +x /opt/lifeos/tg-proxy.py

  cat > /etc/systemd/system/lifeos-tg-proxy.service <<EOF
[Unit]
Description=LifeOS Telegram API proxy via Cloudflare Worker
After=network-online.target

[Service]
ExecStart=/usr/bin/python3 /opt/lifeos/tg-proxy.py ${WORKER_URL} ${PROXY_PORT}
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
