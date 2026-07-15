#!/usr/bin/env bash
# Live log viewer
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd -P)"
LOG="$ROOT/logs/lifeos.log"
mkdir -p "$ROOT/logs"
touch "$LOG"
echo "=== $LOG (Ctrl+C выход) ==="
tail -n 80 -F "$LOG"
