#!/usr/bin/env bash
# Open settings.env in a text editor
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd -P)"
ENV_FILE="$ROOT/settings.env"
EXAMPLE="$ROOT/settings.env.example"
if [[ ! -f "$ENV_FILE" ]]; then
  cp "$EXAMPLE" "$ENV_FILE"
fi
if command -v open >/dev/null 2>&1; then
  open -e "$ENV_FILE" || open -t "$ENV_FILE" || open "$ENV_FILE"
elif command -v xdg-open >/dev/null 2>&1; then
  xdg-open "$ENV_FILE"
elif command -v notepad.exe >/dev/null 2>&1; then
  notepad.exe "$ENV_FILE"
else
  "${EDITOR:-nano}" "$ENV_FILE"
fi
