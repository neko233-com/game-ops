#!/usr/bin/env bash
# Install the game-long-term-operation skill project-first, with a global fallback.
set -euo pipefail

SCOPE="${1:-project}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE="$REPO_ROOT/skills/game-long-term-operation"

if [ ! -f "$SOURCE/SKILL.md" ]; then
  echo "Canonical skill not found at $SOURCE" >&2
  exit 1
fi

find_project_root() {
  dir="$(pwd)"
  while :; do
    for marker in .git go.mod package.json .opencode .claude .agents .dsh; do
      if [ -e "$dir/$marker" ]; then echo "$dir"; return 0; fi
    done
    parent="$(dirname "$dir")"
    [ "$parent" = "$dir" ] && return 0
    dir="$parent"
  done
}

PROJECT_ROOT="$(find_project_root)"
HOME_DIR="${HOME:-$USERPROFILE}"

if [ "$SCOPE" = "project" ] && [ -z "$PROJECT_ROOT" ]; then
  echo "No project detected; falling back to global install." >&2
  SCOPE="global"
fi

install_to() {
  local parent="$1" target="$1/game-long-term-operation" runtime="$2"
  rm -rf "$target"
  mkdir -p "$parent"
  cp -R "$SOURCE" "$target"
  [ -f "$target/SKILL.md" ] || { echo "Install verification failed: $target" >&2; exit 1; }
  echo "  $runtime: $target"
}

echo "Installed game-long-term-operation ($SCOPE):"

case "$SCOPE" in
  project|all)
    if [ -n "$PROJECT_ROOT" ]; then
      install_to "$PROJECT_ROOT/.opencode/skills" opencode
      install_to "$PROJECT_ROOT/.claude/skills" claude
      install_to "$PROJECT_ROOT/.agents/skills" codex
      install_to "$PROJECT_ROOT/.dsh/skills" dsh
    fi
    ;;
esac

case "$SCOPE" in
  global|all)
    install_to "${OPENCODE_SKILLS_DIR:-$HOME_DIR/.config/opencode/skills}" opencode
    install_to "${CLAUDE_SKILLS_DIR:-$HOME_DIR/.claude/skills}" claude
    install_to "${CODEX_SKILLS_DIR:-$HOME_DIR/.agents/skills}" codex
    install_to "${DSH_SKILLS_DIR:-${SKILLS_DIR:-$HOME_DIR/.dsh/skills}}" dsh
    ;;
esac

case "$SCOPE" in
  project|global|all) ;;
  *) echo "Unknown scope: $SCOPE (use project|global|all)" >&2; exit 2 ;;
esac
