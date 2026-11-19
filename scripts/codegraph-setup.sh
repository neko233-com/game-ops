#!/usr/bin/env bash
# Optional, project-level codegraph233 provisioning for flow (心流) analysis.
# NOT covered by this repo's tests. Downloads into <project>/.codegraph233/bin/.
set -euo pipefail

PROJECT_ROOT="${1:-$(pwd)}"
VERSION="${2:-latest}"
REPO='neko233-com/codegraph233'
BIN_DIR="$PROJECT_ROOT/.codegraph233/bin"
EXE="$BIN_DIR/codegraph233"

if command -v codegraph233 >/dev/null 2>&1; then
  EXE="$(command -v codegraph233)"
  echo "Using codegraph233 from PATH: $EXE"
elif [ -x "$EXE" ]; then
  echo "Using project-local codegraph233: $EXE"
else
  echo "Downloading $REPO ($VERSION) into $BIN_DIR ..."
  mkdir -p "$BIN_DIR"
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"
  case "$ARCH" in x86_64|amd64) ARCH='x86_64' ;; aarch64|arm64) ARCH='aarch64' ;; esac
  API="https://api.github.com/repos/$REPO/releases/latest"
  [ "$VERSION" != "latest" ] && API="https://api.github.com/repos/$REPO/releases/tags/$VERSION"
  URL="$(curl -fsSL -H 'User-Agent: game-ops-codegraph-setup' "$API" \
    | grep -o '"browser_download_url": *"[^"]*"' \
    | sed 's/.*: *"//; s/"$//' \
    | grep -i "$OS" | grep -i "$ARCH" | grep -iE '\.(tar\.gz|zip)$' | head -n1 || true)"
  if [ -z "$URL" ]; then
    echo "No matching release asset. Install manually:" >&2
    echo "  git clone https://github.com/$REPO.git && cd codegraph233 && cargo install --path . --force" >&2
    exit 2
  fi
  ARCHIVE="$BIN_DIR/$(basename "$URL")"
  curl -fsSL -H 'User-Agent: game-ops-codegraph-setup' -o "$ARCHIVE" "$URL"
  case "$ARCHIVE" in
    *.zip) unzip -oq "$ARCHIVE" -d "$BIN_DIR" ;;
    *.tar.gz) tar -xzf "$ARCHIVE" -C "$BIN_DIR" ;;
  esac
  rm -f "$ARCHIVE"
  EXE="$(find "$BIN_DIR" -type f -name 'codegraph233' -perm -u+x | head -n1)"
  [ -n "$EXE" ] || { echo "codegraph233 binary not found after extraction" >&2; exit 1; }
  chmod +x "$EXE"
  echo "Downloaded: $EXE"
fi

CONFIG="$PROJECT_ROOT/.codegraph233.json"
if [ ! -f "$CONFIG" ]; then
  cat > "$CONFIG" <<'JSON'
{
  "version": 1,
  "profile": "full",
  "roots": [{ "key": "client", "path": ".", "role": "frontend" }],
  "languages": ["auto"]
}
JSON
  echo "Wrote $CONFIG (engine auto-detection)"
fi

"$EXE" init
"$EXE" index
"$EXE" status

cat <<'EOF'

Next, in your agent, query via codegraph233 MCP:
  cg233_design_graph   - design docs <-> code <-> assets
  cg233_config_tables  - discovered numeric config tables
  cg233_call_graph     - runtime paths / reward loop reachability
  cg233_asset_tree     - content structure supporting flow

Flow analysis is slower than numeric scoring and is NOT tested in this repo.
EOF
