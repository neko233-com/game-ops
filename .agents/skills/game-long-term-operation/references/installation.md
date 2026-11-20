# Install across runtimes (project-first)

The skill follows the Agent Skills spec (a directory with `SKILL.md` + optional
`references/`, `scripts/`, `assets/`). One canonical copy lives in this repo at
`skills/game-long-term-operation/`; installers copy it into each runtime's
skill directory.

## Project-level priority (default)

Project-level skills override global ones and are shared with teammates through
the repo. The installer walks up from the current directory looking for a project
marker (`.git`, `go.mod`, `package.json`, `.opencode`, `.claude`, `.agents`,
`.dsh`) and installs there first. Only when no project is detected does it fall
back to the global user directory.

| Runtime | Project dir | Global dir |
|---|---|---|
| OpenCode | `<project>/.opencode/skills/game-long-term-operation/` | `~/.config/opencode/skills/game-long-term-operation/` |
| Claude Code | `<project>/.claude/skills/game-long-term-operation/` | `~/.claude/skills/game-long-term-operation/` |
| Codex / agentskills alias | `<project>/.agents/skills/game-long-term-operation/` | `~/.agents/skills/game-long-term-operation/` |
| dsh | `<project>/.dsh/skills/game-long-term-operation/` | `~/.dsh/skills/game-long-term-operation/` (or `$DSH_SKILLS_DIR`) |

Environment overrides apply to **global** targets only (e.g. `OPENCODE_SKILLS_DIR`,
`CODEX_SKILLS_DIR`, `DSH_SKILLS_DIR`, with generic `SKILLS_DIR` as the dsh
fallback). Project targets are always the fixed paths in the table above and are
not redirected. The installer prints every path it writes.

## One-command install

After cloning this repo and reading its root `README.md`:

```sh
# macOS / Linux
scripts/install-skill.sh
scripts/install-skill.sh --scope global
scripts/install-skill.sh --scope all
```

```powershell
# Windows
scripts/install-skill.ps1
scripts/install-skill.ps1 -Scope global
scripts/install-skill.ps1 -Scope all
```

The script:
1. Detects the project root (project-first).
2. Copies `skills/game-long-term-operation/` into each applicable target.
3. Verifies each target has `SKILL.md`.
4. Prints a summary of installed paths.

## Manual install

Copy the folder yourself, e.g. for OpenCode project scope:

```sh
cp -r skills/game-long-term-operation <project>/.opencode/skills/
```

Or global Codex:

```sh
cp -r skills/game-long-term-operation ~/.agents/skills/
```

## Verify

- The runtime should list `game-long-term-operation` among available skills.
- Ask the agent: "Use the game-long-term-operation skill to audit my combat
  curve." It should load `SKILL.md` and follow the canonical loop.
- `SKILL.md` frontmatter must keep `name: game-long-term-operation` (it must
  match the directory name) and a `description` starting with "Use when...".

## Updating

Re-run the installer after pulling; it overwrites the copied skill with the
canonical version. Keep the canonical repo copy as the single source of truth.
