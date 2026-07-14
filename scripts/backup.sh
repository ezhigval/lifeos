#!/usr/bin/env bash
set -euo pipefail

: "${LIFEOS_DATABASE_URL:?set LIFEOS_DATABASE_URL}"
OUT_DIR="${1:-./backups}"
mkdir -p "$OUT_DIR"
FILE="$OUT_DIR/lifeos_$(date +%Y%m%d_%H%M%S).dump"
pg_dump "$LIFEOS_DATABASE_URL" -Fc -f "$FILE"
echo "backup saved: $FILE"
