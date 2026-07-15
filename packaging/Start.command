#!/usr/bin/env bash
# Start LifeOS: migrate → serve, stream logs to Terminal + logs/lifeos.log
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd -P)"
# shellcheck disable=SC1091
source "$ROOT/lib.sh"

lifeos_load_env "$ROOT"
BIN="$(lifeos_bin "$ROOT")"
LOG="$ROOT/logs/lifeos.log"
PIDFILE="$ROOT/run/lifeos.pid"

cd "$ROOT"

if [[ -f "$PIDFILE" ]] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
  echo "LifeOS уже запущен (pid $(cat "$PIDFILE"))."
  echo "Логи: ./Logs.command   Стоп: ./Stop.command"
  exec "$ROOT/Logs.command"
fi

echo "==> $($BIN version 2>/dev/null | head -1 || echo lifeos)"
echo "==> migrate up…"
"$BIN" migrate up

echo "==> serve ${LIFEOS_HTTP_ADDR:-:8080}"
echo "    log file: $LOG"
echo "    Ctrl+C останавливает сервер"
echo

"$BIN" serve >>"$LOG" 2>&1 &
SERVE_PID=$!
echo "$SERVE_PID" >"$PIDFILE"

cleanup() {
  if kill -0 "$SERVE_PID" 2>/dev/null; then
    kill "$SERVE_PID" 2>/dev/null || true
    wait "$SERVE_PID" 2>/dev/null || true
  fi
  rm -f "$PIDFILE"
}
trap cleanup EXIT INT TERM

# Show last lines then follow (same Terminal = one double-click UX).
tail -n 5 -F "$LOG" &
TAIL_PID=$!
wait "$SERVE_PID"
kill "$TAIL_PID" 2>/dev/null || true
