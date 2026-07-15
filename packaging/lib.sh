#!/usr/bin/env bash
# Shared helpers for LifeOS desktop package launchers.
# shellcheck disable=SC1091

set -euo pipefail

lifeos_root() {
  local src="${BASH_SOURCE[1]:-${BASH_SOURCE[0]}}"
  cd "$(dirname "$src")" && pwd -P
}

lifeos_load_env() {
  local root="$1"
  local env_file="$root/settings.env"
  local example="$root/settings.env.example"

  if [[ ! -f "$env_file" ]]; then
    cp "$example" "$env_file"
    echo "Создан settings.env — заполни TELEGRAM_BOT_TOKEN, JWT, DATABASE_URL."
    if command -v open >/dev/null 2>&1; then
      open -e "$env_file" || open "$env_file" || true
    elif command -v xdg-open >/dev/null 2>&1; then
      xdg-open "$env_file" >/dev/null 2>&1 || true
    fi
    echo "Сохрани файл и запусти Start снова."
    exit 0
  fi

  set -a
  # shellcheck disable=SC1090
  source "$env_file"
  set +a

  export LIFEOS_STATIC_DIR="${LIFEOS_STATIC_DIR:-$root/web/miniapp/dist}"
  export LIFEOS_MIGRATIONS_DIR="${LIFEOS_MIGRATIONS_DIR:-$root/migrations}"
  mkdir -p "$root/logs" "$root/run"
}

lifeos_bin() {
  local root="$1"
  if [[ -x "$root/bin/lifeos" ]]; then
    echo "$root/bin/lifeos"
  elif [[ -x "$root/bin/lifeos.exe" ]]; then
    echo "$root/bin/lifeos.exe"
  else
    echo "Нет бинарника bin/lifeos — пересобери пакет (make package)." >&2
    exit 1
  fi
}
