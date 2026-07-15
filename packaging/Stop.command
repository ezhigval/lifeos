#!/usr/bin/env bash
# Stop LifeOS if started via Start.command
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd -P)"
PIDFILE="$ROOT/run/lifeos.pid"

if [[ ! -f "$PIDFILE" ]]; then
  echo "PID-файл не найден — сервер не из этого пакета или уже остановлен."
  # best-effort: kill by name in this package path
  pkill -f "$ROOT/bin/lifeos serve" 2>/dev/null || true
  exit 0
fi

PID="$(cat "$PIDFILE")"
if kill -0 "$PID" 2>/dev/null; then
  kill "$PID" 2>/dev/null || true
  sleep 0.5
  kill -9 "$PID" 2>/dev/null || true
  echo "Остановлен pid $PID"
else
  echo "Процесс $PID уже мёртв"
fi
rm -f "$PIDFILE"
