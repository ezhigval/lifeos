#!/usr/bin/env python3
"""Local forward proxy: absolute-form http://api.telegram.org/... → Worker /fetch.

The Go client must use plain HTTP when LIFEOS_HTTP_PROXY is set. HTTPS would
CONNECT to api.telegram.org and never hit this process.

Env (preferred, so the secret is not on the command line):
  LIFEOS_TG_PROXY_WORKER_URL   https://<name>.<account>.workers.dev
  LIFEOS_TG_PROXY_SECRET       optional, sent as x-proxy-secret
  TG_PROXY_PORT                default 8081
  TG_PROXY_BIND                comma-separated bind addresses, default 127.0.0.1

argv[1] is accepted as the worker URL only when the env var is empty.
"""
import json
import os
import re
import sys
import threading
import urllib.error
import urllib.parse
import urllib.request
from http.server import BaseHTTPRequestHandler
from socketserver import ThreadingTCPServer

WORKER = os.environ.get("LIFEOS_TG_PROXY_WORKER_URL", "").strip().rstrip("/")
if not WORKER and len(sys.argv) > 1 and not sys.argv[1].startswith("-"):
    WORKER = sys.argv[1].strip().rstrip("/")
PORT = int(os.environ.get("TG_PROXY_PORT", sys.argv[2] if len(sys.argv) > 2 else "8081"))
SECRET = os.environ.get("LIFEOS_TG_PROXY_SECRET", "").strip()
BINDS = [b.strip() for b in os.environ.get("TG_PROXY_BIND", "127.0.0.1").split(",") if b.strip()]

_PATH_RE = re.compile(r"^https?://api\.telegram\.org(/.*)$")


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def _target(self):
        match = _PATH_RE.match(self.path)
        return match.group(1) if match else None

    def _relay(self, method):
        target = self._target()
        if not target or ".." in target:
            self.send_error(404)
            return
        body = None
        if method not in ("GET", "HEAD"):
            n = int(self.headers.get("Content-Length", 0))
            body = self.rfile.read(n) if n else None
        headers = {}
        incoming = self.headers.get("Content-Type")
        if incoming:
            headers["Content-Type"] = incoming
        elif body is not None:
            headers["Content-Type"] = "application/json"
        if SECRET:
            headers["x-proxy-secret"] = SECRET
        # Quote the whole path so '&' inside getUpdates does not break the worker query.
        quoted = urllib.parse.quote(target, safe="")
        req = urllib.request.Request(
            WORKER + "/fetch?path=" + quoted,
            data=body,
            method=method,
            headers=headers,
        )
        try:
            with urllib.request.urlopen(req, timeout=60) as resp:
                data, code = resp.read(), resp.status
                ctype = resp.headers.get("Content-Type") or "application/octet-stream"
        except urllib.error.HTTPError as exc:
            data, code = exc.read(), exc.code
            ctype = exc.headers.get("Content-Type") if exc.headers else None
            ctype = ctype or "application/octet-stream"
        except Exception:
            # Do not stringify the exception: urllib includes the request URL,
            # and that URL contains the bot token.
            data = json.dumps({"ok": False, "error": "upstream unreachable"}).encode()
            code = 502
            ctype = "application/json"
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_POST(self):  # noqa: N802
        self._relay("POST")

    def do_GET(self):  # noqa: N802
        self._relay("GET")

    def log_message(self, *_args):
        return


def _serve(addr: str) -> None:
    ThreadingTCPServer.allow_reuse_address = True
    try:
        srv = ThreadingTCPServer((addr, PORT), Handler)
    except OSError as exc:
        sys.stderr.write(f"tg-proxy bind {addr}:{PORT} failed: {exc.errno}\n")
        sys.exit(1)
    srv.serve_forever()


def main() -> None:
    if not WORKER:
        sys.stderr.write("LIFEOS_TG_PROXY_WORKER_URL is required\n")
        sys.exit(1)
    if not BINDS:
        sys.stderr.write("TG_PROXY_BIND is empty\n")
        sys.exit(1)
    sys.stderr.write(f"tg-proxy listening on {','.join(BINDS)}:{PORT}\n")
    threads = []
    for addr in BINDS[:-1]:
        thread = threading.Thread(target=_serve, args=(addr,), daemon=True)
        thread.start()
        threads.append(thread)
    _serve(BINDS[-1])


if __name__ == "__main__":
    main()
