#!/usr/bin/env bash
# Build production release of the single `gameops` binary for multiple platforms.
# Produces dist/<version>/gameops-<os>-<arch>[.exe] (flat, GitHub-Release-ready)
# + SHA256SUMS (matching asset names) + version.json.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

VERSION="${VERSION:-v1.0.0}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo none)}"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
OUT="dist/$VERSION"
LDFLAGS="-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"

TARGETS=(
  "windows/amd64" "windows/arm64"
  "linux/amd64"   "linux/arm64"
  "darwin/amd64"  "darwin/arm64"
)

echo "Building gameops $VERSION ($COMMIT) ..."
rm -rf "$OUT"
mkdir -p "$OUT"

for target in "${TARGETS[@]}"; do
  GOOS="${target%/*}"
  GOARCH="${target#*/}"
  ext=""
  [ "$GOOS" = "windows" ] && ext=".exe"
  bin="gameops-$GOOS-$GOARCH$ext"
  echo "  -> $bin"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
    go build -trimpath -ldflags "$LDFLAGS" -o "$OUT/$bin" ./cmd/gameops
done

cd "$OUT"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum gameops-* > SHA256SUMS
elif command -v shasum >/dev/null 2>&1; then
  shasum -a 256 gameops-* > SHA256SUMS
else
  echo "no sha256 tool found" >&2
  exit 1
fi

cat > version.json <<EOF
{
  "version": "$VERSION",
  "commit": "$COMMIT",
  "date": "$DATE",
  "go": "$(go version | awk '{print $3}')",
  "targets": ["windows-amd64", "windows-arm64", "linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64"]
}
EOF

echo "Done: $OUT"
echo "Verify: cd $OUT && sha256sum -c SHA256SUMS"
