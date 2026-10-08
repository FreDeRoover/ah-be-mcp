#!/usr/bin/env bash
# Build release archives and Claude Desktop .mcpb bundles into dist/.
# Usage: scripts/release.sh 1.2.3
set -euo pipefail
v="${1:?usage: scripts/release.sh <version, e.g. 1.2.3>}"
cd "$(dirname "$0")/.."
rm -rf dist && mkdir dist

# goos goarch label mcpb-platform (empty = no .mcpb, Claude Desktop is macOS/Windows only)
while read -r goos goarch label mplat; do
  ext=""; [ "$goos" = windows ] && ext=".exe"
  stage="dist/stage-$label"; mkdir -p "$stage/server"
  GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$v" \
    -o "$stage/server/ah-be-mcp$ext" .

  # plain archive: binary + README + LICENSE
  a="ah-be-mcp_${v}_${label}"; mkdir -p "dist/$a"
  cp "$stage/server/ah-be-mcp$ext" README.md LICENSE "dist/$a/"
  (cd dist && if [ "$goos" = windows ]; then zip -qr "$a.zip" "$a"; else tar czf "$a.tar.gz" "$a"; fi; rm -rf "$a")

  # Claude Desktop bundle
  if [ -n "$mplat" ]; then
    sed -e "s/__VERSION__/$v/" -e "s/__PLATFORM__/$mplat/" mcpb/manifest.json > "$stage/manifest.json"
    (cd "$stage" && zip -qr "../ah-be-mcp_${v}_${label}.mcpb" manifest.json server)
  fi
  rm -rf "$stage"
done <<LIST
darwin arm64 macos-apple-silicon darwin
darwin amd64 macos-intel darwin
windows amd64 windows win32
linux amd64 linux-amd64
linux arm64 linux-arm64
LIST

(cd dist && shasum -a 256 * > checksums.txt)
ls -l dist
