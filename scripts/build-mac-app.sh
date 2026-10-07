#!/usr/bin/env bash
# Build LifeOS.app for Apple Silicon and Intel. Swift compiles only on macOS.
# The Go server and the UI are produced everywhere; the .app zip is produced on a Mac.
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
cd "$root"

npm --prefix web/miniapp ci
npm --prefix web/miniapp run build:desktop

ui="$root/cmd/lifeos-desktop/ui"
rm -rf "$ui"
mkdir -p "$ui"
cp -R "$root/web/miniapp/desktop/dist/." "$ui/"

mkdir -p "$root/dist/mac"
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o "$root/dist/mac/lifeos-desktop-arm64" ./cmd/lifeos-desktop
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$root/dist/mac/lifeos-desktop-amd64" ./cmd/lifeos-desktop

if [[ "$(uname)" != "Darwin" ]]; then
  echo "Go binaries are in dist/mac. The .app bundle is assembled on macOS."
  exit 0
fi

swiftc -target arm64-apple-macosx13.0 -O -o "$root/dist/mac/LifeOS-arm64" "$root/desktop/macos/main.swift" -framework Cocoa -framework WebKit
swiftc -target x86_64-apple-macosx13.0 -O -o "$root/dist/mac/LifeOS-amd64" "$root/desktop/macos/main.swift" -framework Cocoa -framework WebKit

app="$root/dist/LifeOS.app"
rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
cp "$root/desktop/macos/Info.plist" "$app/Contents/Info.plist"
lipo -create -output "$app/Contents/MacOS/LifeOS" "$root/dist/mac/LifeOS-arm64" "$root/dist/mac/LifeOS-amd64"
lipo -create -output "$app/Contents/MacOS/lifeos-desktop" "$root/dist/mac/lifeos-desktop-arm64" "$root/dist/mac/lifeos-desktop-amd64"
chmod +x "$app/Contents/MacOS/LifeOS" "$app/Contents/MacOS/lifeos-desktop"
codesign --force --deep --sign - "$app"
rm -f "$root/dist/LifeOS-mac.zip"
ditto -c -k --keepParent "$app" "$root/dist/LifeOS-mac.zip"
echo "wrote $root/dist/LifeOS-mac.zip"
