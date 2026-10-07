#!/usr/bin/env python3
"""Local forward proxy: absolute-form http://api.telegram.org/... → Worker /fetch.

The Go client must use plain HTTP when LIFEOS_HTTP_PROXY is set. HTTPS would
CONNECT to api.telegram.org and never hit this process.

Env (preferred, so the secret is not on the command line):
  LIFEOS_TG_PROXY_WORKER_URL   https://host that serves /fetch  (apex or workers.dev)
  LIFEOS_TG_PROXY_EDGE_IPS     optional comma-separated IPv4s to dial instead of DNS
  LIFEOS_TG_PROXY_SECRET       optional, sent as x-proxy-secret
  TG_PROXY_PORT                default 8081
  TG_PROXY_BIND                comma-separated bind addresses, default 127.0.0.1

Edge IPs are for networks that can open some Cloudflare edges but not the
anycast addresses DNS returns (Yandex blocks 104.21/172.67 and Telegram).
The TLS SNI and Host header stay on the worker URL. An edge that answers
with Cloudflare 1034 or the origin text "404 page not found" is skipped.

argv[1] is accepted as the worker URL only when the env var is empty.
"""
import http.client
import json
import os
import re
import socket
import ssl
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
EDGE = [ip.strip() for ip in os.environ.get("LIFEOS_TG_PROXY_EDGE_IPS", "").split(",") if ip.strip()]
_PINNED = {"ip": ""}

_PATH_RE = re.compile(r"^https?://api\.telegram\.org(/.*)$")
_TLS = ssl.create_default_context()
# Cloudflare edges prefer HTTP/2 when ALPN offers it. This proxy speaks HTTP/1.1.
try:
    _TLS.set_alpn_protocols(["http/1.1"])
except Exception:
    pass


def _edge_order():
    pinned = _PINNED["ip"]
    if pinned and pinned in EDGE:
        return [pinned] + [ip for ip in EDGE if ip != pinned]
    return list(EDGE)


def _skip_edge(code, data):
    """True when this edge did not run the worker."""
    if code == 403 and b"error code: 1034" in data:
        return True
    if code == 404 and data.strip() == b"404 page not found":
        return True
    return False


def _dial(ip, host, port, method, path, body, headers):
    conn = http.client.HTTPSConnection(host, port, timeout=20, context=_TLS)
    try:
        raw = socket.create_connection((ip, port), 5)
        raw.settimeout(12)
        conn.sock = _TLS.wrap_socket(raw, server_hostname=host)
        conn.request(method, path, body=body, headers=headers)
        resp = conn.getresponse()
        data = resp.read()
        ctype = resp.getheader("Content-Type") or "application/octet-stream"
        return resp.status, data, ctype
    finally:
        conn.close()


def _via_edges(method, path, body, headers):
    parsed = urllib.parse.urlsplit(WORKER)
    host = parsed.hostname
    if not host:
        raise OSError("worker url has no host")
    port = parsed.port or 443
    for ip in _edge_order():
        try:
            code, data, ctype = _dial(ip, host, port, method, path, body, headers)
        except OSError:
            continue
        if _skip_edge(code, data):
            continue
        _PINNED["ip"] = ip
        return code, data, ctype
    raise OSError("no edge served the worker")


def _via_urllib(method, path, body, headers):
    req = urllib.request.Request(WORKER + path, data=body, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            ctype = resp.headers.get("Content-Type") or "application/octet-stream"
            return resp.status, resp.read(), ctype
    except urllib.error.HTTPError as exc:
        ctype = exc.headers.get("Content-Type") if exc.headers else None
        return exc.code, exc.read(), ctype or "application/octet-stream"


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
        fetch_path = "/fetch?path=" + quoted
        try:
            if EDGE:
                code, data, ctype = _via_edges(method, fetch_path, body, headers)
            else:
                code, data, ctype = _via_urllib(method, fetch_path, body, headers)
        except Exception:
            # Do not stringify the exception: it can include the request URL,
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
