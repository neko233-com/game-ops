#!/usr/bin/env bash
# Build production release of the single `gameops` binary for multiple platforms.
# Produces dist/<version>/<os>-<arch>/gameops[.exe] + SHA256SUMS + version.json.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

VERSION="${VERSION:-v1.0.0}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo none)}"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
OUT="dist/$VERSION"
LDFLAGS="-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"

TARGETS=(
  "windows/amd64/gameops.exe"
  "windows/arm64/gameops.exe"
  "linux/amd64/gameops"
  "linux/arm64/gameops"
  "darwin/amd64/gameops"
  "darwin/arm64/gameops"
)

echo "Building gameops $VERSION ($COMMIT) ..."
rm -rf "$OUT"
mkdir -p "$OUT"

for target in "${TARGETS[@]}"; do
  GOOS="${target%%/*}"
  rest="${target#*/}"
  GOARCH="${rest%%/*}"
  bin="${rest#*/}"
  dir="$OUT/$GOOS-$GOARCH"
  mkdir -p "$dir"
  echo "  -> $GOOS/$GOARCH"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
    go build -trimpath -ldflags "$LDFLAGS" -o "$dir/$bin" ./cmd/gameops
done

# SHA256 checksums (portable: sha256sum or shasum)
cd "$OUT"
if command -v sha256sum >/dev/null 2>&1; then
  find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > SHA256SUMS
elif command -v shasum >/dev/null 2>&1; then
  find . -type f ! -name SHA256SUMS -print | sort | xargs shasum -a 256 > SHA256SUMS
else
  echo "no sha256 tool found" >&2
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
