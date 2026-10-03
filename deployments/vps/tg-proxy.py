#!/usr/bin/env python3
"""Minimal HTTP forward proxy that tunnels api.telegram.org via a Cloudflare Worker.

Used on Yandex Cloud VMs where direct egress to Telegram is blocked, but
Cloudflare edge (workers.dev) is reachable. The Go app points LIFEOS_HTTP_PROXY
at this proxy; requests to http://api.telegram.org/... are rewritten into
<worker>/fetch?path=/bot<token>/<method> and replayed by the Worker.

Usage: tg-proxy.py <https://tg-proxy.<sub>.workers.dev> [port]
"""
import http.server
import json
import re
import socketserver
import sys
import urllib.request

WORKER = sys.argv[1].rstrip("/")
PORT = int(sys.argv[2]) if len(sys.argv) > 2 else 8081

_PATH_RE = re.compile(r"^http://api\.telegram\.org(/.*)$")


class Handler(http.server.BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def _target(self):
        m = _PATH_RE.match(self.path)
        return m.group(1) if m else None

    def _relay(self, method):
        t = self._target()
        if not t:
            self.send_error(404)
            return
        body = None
        if method == "POST":
            n = int(self.headers.get("Content-Length", 0))
            body = self.rfile.read(n)
        req = urllib.request.Request(
            WORKER + "/fetch?path=" + t,  # keep leading slash: /bot<token>/<method>
            data=body,
            method=method,
            headers={"Content-Type": "application/json"},
        )
        try:
            with urllib.request.urlopen(req, timeout=30) as r:
                data, code = r.read(), r.status
        except Exception as e:  # noqa: BLE001
            data = json.dumps({"ok": False, "error": str(e)}).encode()
            code = 502
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_POST(self):  # noqa: N802
        self._relay("POST")

    def do_GET(self):  # noqa: N802
        self._relay("GET")

    def log_message(self, *args):  # silence
        pass


def main():
    socketserver.ThreadingTCPServer.allow_reuse_address = True
    with socketserver.ThreadingTCPServer(("127.0.0.1", PORT), Handler) as srv:
        srv.serve_forever()


if __name__ == "__main__":
    main()
