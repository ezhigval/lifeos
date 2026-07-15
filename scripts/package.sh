#!/usr/bin/env bash
# Build a click-to-run LifeOS package for one GOOS/GOARCH.
# Usage:
#   ./scripts/package.sh                 # host OS/arch
#   ./scripts/package.sh darwin arm64    # Mac Apple Silicon
#   ./scripts/package.sh windows amd64
#   ./scripts/package.sh linux amd64
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd -P)"
cd "$ROOT"

GOOS="${1:-$(go env GOOS)}"
GOARCH="${2:-$(go env GOARCH)}"
VERSION="${LIFEOS_VERSION:-LifeOS_alpha_1.0.0}"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
BUILT_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
PKG_NAME="${VERSION}_${GOOS}_${GOARCH}"
OUT="${ROOT}/dist/${PKG_NAME}"
BIN_NAME="lifeos"
[[ "$GOOS" == "windows" ]] && BIN_NAME="lifeos.exe"

echo "==> package $PKG_NAME"
rm -rf "$OUT"
mkdir -p "$OUT/bin" "$OUT/logs" "$OUT/run" "$OUT/web/miniapp" "$OUT/migrations"

echo "==> miniapp"
if [[ ! -d web/miniapp/dist ]] || [[ "${FORCE_MINIAPP_BUILD:-}" == "1" ]]; then
  (cd web/miniapp && npm ci && npm run build)
fi
cp -R web/miniapp/dist "$OUT/web/miniapp/"

echo "==> migrations"
cp -R migrations/. "$OUT/migrations/"

echo "==> binary ($GOOS/$GOARCH)"
LDFLAGS=(
  -X "github.com/valentinezhov/lifeos/cmd/lifeos/cmd.Version=${VERSION}"
  -X "github.com/valentinezhov/lifeos/cmd/lifeos/cmd.Commit=${COMMIT}"
  -X "github.com/valentinezhov/lifeos/cmd/lifeos/cmd.BuiltAt=${BUILT_AT}"
)
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
  -trimpath \
  -ldflags "${LDFLAGS[*]}" \
  -o "$OUT/bin/$BIN_NAME" \
  ./cmd/lifeos

echo "==> launchers + settings"
cp packaging/settings.env.example "$OUT/settings.env.example"
cp packaging/README.txt "$OUT/README.txt"
cp packaging/lib.sh "$OUT/lib.sh"
cp packaging/Start.command packaging/Stop.command packaging/Logs.command packaging/Settings.command "$OUT/"
cp packaging/start.sh "$OUT/"
cp packaging/Start.bat packaging/Stop.bat packaging/Logs.bat packaging/Settings.bat "$OUT/"
chmod +x "$OUT"/Start.command "$OUT"/Stop.command "$OUT"/Logs.command "$OUT"/Settings.command \
  "$OUT"/start.sh "$OUT"/lib.sh "$OUT/bin/$BIN_NAME" 2>/dev/null || true

# First-run settings copy is created by Start; keep example only in zip.

ARCHIVE="${ROOT}/dist/${PKG_NAME}.tar.gz"
(
  cd "${ROOT}/dist"
  tar -czf "${PKG_NAME}.tar.gz" "${PKG_NAME}"
)
if command -v zip >/dev/null 2>&1; then
  (
    cd "${ROOT}/dist"
    rm -f "${PKG_NAME}.zip"
    zip -qr "${PKG_NAME}.zip" "${PKG_NAME}"
  )
  echo "==> zip  dist/${PKG_NAME}.zip"
fi

echo "==> done"
echo "    folder: $OUT"
echo "    archive: $ARCHIVE"
if [[ "$GOOS" == "$(go env GOOS)" && "$GOARCH" == "$(go env GOARCH)" ]]; then
  "$OUT/bin/$BIN_NAME" version || true
fi
