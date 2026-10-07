#!/usr/bin/env bash
# Build the Windows desktop window zip: lifeos-desktop.exe plus install.bat.
# The exe is the same local UI server the Mac app embeds. Edge or Chrome
# opens it as a window (LifeOS.bat). This script cross-compiles on Linux.
# It does not produce a signed install.exe, and the desktop-release workflow
# does not run it: that workflow only has a macOS runner.
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
cd "$root"
raw="${LIFEOS_DESKTOP_VERSION:-desktop-v0.2.0}"
version="${raw#desktop-v}"
case "$version" in
  ''|*[!0-9.]*) echo "bad version: $raw" >&2; exit 1 ;;
esac

npm --prefix web/miniapp ci
npm --prefix web/miniapp run build:desktop

ui="$root/cmd/lifeos-desktop/ui"
rm -rf "$ui"
mkdir -p "$ui"
cp -R "$root/web/miniapp/desktop/dist/." "$ui/"
printf '*\n!index.html\n!.gitignore\n' > "$ui/.gitignore"

mkdir -p "$root/dist/win"
ldflags="-s -w -H windowsgui -X main.desktopVersion=${version}"
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$ldflags" -o "$root/dist/win/lifeos-desktop.exe" ./cmd/lifeos-desktop

stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT
cp "$root/dist/win/lifeos-desktop.exe" "$stage/lifeos-desktop.exe"
for f in LifeOS.bat install.bat README.txt; do
  sed 's/$/\r/' "$root/packaging/desktop-win/$f" > "$stage/$f"
done
mkdir -p "$root/dist"
rm -f "$root/dist/LifeOS-win.zip"
(
  cd "$stage"
  python3 -c 'import os, sys, zipfile; z=zipfile.ZipFile(sys.argv[1], "w", zipfile.ZIP_DEFLATED); [z.write(n, n) for n in sorted(os.listdir("."))]; z.close()' "$root/dist/LifeOS-win.zip"
)
echo "wrote $root/dist/LifeOS-win.zip"
