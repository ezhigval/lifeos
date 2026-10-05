#!/bin/sh
# Apply schema before the long-running server. One-off commands
# (migrate status, version) are passed through unchanged.
set -eu
if [ "${1:-}" = "serve" ]; then
  /app/lifeos migrate up
fi
exec /app/lifeos "$@"
