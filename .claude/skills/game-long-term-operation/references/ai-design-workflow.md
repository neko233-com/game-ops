# AI designs all numbers; million-file analysis; flow

## The mandate: the agent owns every number

The agent is not limited to *reviewing* numbers. It designs **all** of them:

- **战斗数值投放** — damage, HP, defense, skill coefficients, drop tables.
- **成长数值** — XP/level curves, attribute growth, cap placement, prestige.
- **商业化数值** — prices, reward density, pay-rate/LTV targets, bundle value.

For each it writes data-driven curves (never magic numbers in logic), runs the
validator, and explains every choice in the value/flow/pacing language of the
unified theory. Human taste can override **data parameters** (curve constants, band widths,
target values), but the agent produces the complete first draft and the exact
rationale. "Human override" never covers the method itself: no float64, no
global exponent, no skipped validation, no committed test files.

### Design procedure

1. Declare the base unit and the exchange-rate ledger (`normalize`).
2. Choose curve parameters: `base`, `step`, `cap` for each curve — all additive.
3. Set `maxGrowthRatio` (e.g. `1.5`) as the explosion guard.
4. Generate tables with `cmd/numgen`; inspect `issues` from `Plan.Validate`.
5. Normalize all rewards/prices; check source/sink ratio ≈ 1.
6. Write configs as JSON with **string** numbers; re-run ingest to confirm.
7. Check the economy with `cmd/econcheck` (`examples/econ.json`):
   `sources` / `sinks` maps of string values → exact ratio, target ≈ 1.
8. Produce a one-page rationale: what changed, in base units, and why pacing improved.

## Streaming 1,000,000 config files

`ingest.Scan` is built for scale and bounded memory:

- `filepath.WalkDir` produces paths lazily; a bounded worker pool processes them.
- A channel with a fixed capacity means at most `workers` files are open at once.
- Formats: `.xlsx` (stdlib `archive/zip` + `encoding/xml` — no cgo, no external
  deps), `.json` / `.jsonl` (streaming `json.Decoder`), `.csv`, `.tsv`.
- Per numeric column it accumulates exact `Count/Min/Max/Sum` with `num.Num`.
- `MaxFiles` caps the run (default ceiling 1,000,000); `ctx` cancellation works.
- Bad files are counted in `Errors` and skipped — one corrupt xlsx must not abort
  a million-file audit.

```sh
go run ./cmd/ingest -roots ./Config,./Tables -max 1000000 -workers 16 > stats.json
```

## Letting AI judge the current build

Give the agent exactly three inputs:

1. **The design plan** — intended curves, caps, base unit (from `cmd/numgen`).
2. **The observed stats** — `stats.json` from `cmd/ingest`, including outliers
   (column min/max far from the intended band).
3. **The health report** — `ops.Evaluate` score, verdict and diagnostics.

Ask for: a verdict (healthy / attention / critical), the top 3 deviations in
base units, and concrete config edits (field, old string, new string). Then
re-run `numgen`, `ingest`, and `ops` to confirm the fix. Never let the agent edit
the report by hand — it must edit source configs and re-measure.

Prompt skeleton:

> You own all numbers. Base unit: <unit>. Intended plan: <plan>. Observed stats:
> <stats>. Health: <report>. Judge this build. List the worst 3 value/flow/pacing
> problems with exact numbers, propose config edits as string values, and state
> what you expect each metric to become after the change.

## Flow (心流) analysis via codegraph233 — optional / slower / NOT tested here

codegraph233 (`https://github.com/neko233-com/codegraph233`) is a Rust MIT
GraphRAG MCP server for game projects. It indexes code (C#/Go/Rust/TS/Lua/...),
Unity/Unreal/Godot assets, prefab/scene/shader, and **config tables**, exposing
35 MCP tools. It is the right tool for the qualitative layer numbers cannot see:
how systems connect, whether content layout supports a flow state, whether a
reward loop is actually reachable from code. Discover the current tool list at
runtime with `codegraph233 languages --json` and the MCP tools listing — do not
assume the five tools named below are exhaustive.

### Auto-provision at project level

The agent should run the bundled setup script on demand. It:

1. Checks `codegraph233` on PATH; if absent, downloads the matching release
   binary into `<project>/.codegraph233/bin/` (project-level, not global).
2. Writes a minimal `.codegraph233.json` with engine auto-detection.
3. Runs `init`, `index`, `status` (incremental afterwards).
4. Queries via MCP: `cg233_design_graph`, `cg233_config_tables`,
   `cg233_call_graph`, `cg233_asset_tree`, `cg233_shader_info`.

```
scripts/codegraph-setup.ps1 -Version v0.4.1   # Windows, PINNED
scripts/codegraph-setup.sh . v0.4.1           # macOS/Linux, PINNED
```

The script downloads into `<project>/.codegraph233/bin/` and does **not** modify
PATH, so invoke the binary by its explicit path. Minimal `.codegraph233.json`:

```json
{
  "version": 1,
  "profile": "full",
  "roots": [{ "key": "client", "path": ".", "role": "frontend" }],
  "languages": ["auto"]
}
```

MCP wiring (stdio) — add to the runtime's MCP config, pointing at the project-local
binary (it is then started automatically; bare `codegraph233` commands are only
for manual verification):

```json
{
  "mcpServers": {
    "codegraph233": {
      "command": "<project>/.codegraph233/bin/codegraph233",
      "args": ["serve"]
    }
  }
}
```

Useful tools after indexing: `cg233_design_graph`, `cg233_config_tables`,
`cg233_call_graph`, `cg233_asset_tree`, `cg233_shader_info`.

### Honest limitations

- Flow analysis is much slower than numeric scoring; it is a separate, optional
  pass. Never block a numbers task on it.
- This integration is deliberately **not covered by this repo's tests**; the
  scripts are provided as a starting point and should be validated in the real
  game project.
- codegraph233 is MIT-licensed and maintained separately; pin the version with
  `-Version vX.Y.Z` so a release change cannot alter flow conclusions silently.
- MCP tools return static graph facts, not opinions — the agent still judges.
