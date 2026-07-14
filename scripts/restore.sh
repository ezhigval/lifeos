#!/usr/bin/env bash
set -euo pipefail

: "${LIFEOS_DATABASE_URL:?set LIFEOS_DATABASE_URL}"
DUMP_FILE="${1:?usage: ./scripts/restore.sh <dump-file>}"

if [[ ! -f "$DUMP_FILE" ]]; then
  echo "dump file not found: $DUMP_FILE" >&2
  exit 1
fi

pg_restore --clean --if-exists --no-owner --dbname="$LIFEOS_DATABASE_URL" "$DUMP_FILE"
echo "restore completed from: $DUMP_FILE"
