---
name: game-long-term-operation
description: Use when designing, auditing, balancing or monetizing long-term game operations (游戏长线运营) and ALL numerical design (数值) — unit value & normalization (归一化/单位价值), additive-first formulas (加法优先), combat/growth/monetization numbers, an agent designing every number, streaming analysis of 1,000,000 Excel/JSON/CSV config files, judging a build's numerical health, predicting game lifespan from numbers, and optional codegraph233 flow (心流) analysis. Genre-agnostic, exact math, project-first install for opencode/codex/dsh/claude.
---

# Game Long-Term Operation & Numerical Design (游戏长线运营 / 数值)

## The one principle (大道归一)

> 数值的本质不是公式，而是**「价值的定义、流动与节奏」**。
> Numbers are not formulas: they are the definition of value, the flow of value,
> and the pacing of cost versus reward. Everything else is a special case.

Three questions reduce every numerical task:

1. **定义** — what is the base unit of value, and what is everything else worth in it?
2. **流动** — where does value enter (source) and leave (sink)? Is the flow balanced?
3. **节奏** — how fast is value produced, consumed and replaced over time?

If you can answer these with exact numbers, you can design combat, growth and
monetization with the same method, for any genre. Full essay: `references/unified-theory.md`.

**Can game lifespan be judged from numbers?** Partly — numbers are a leading
indicator, not a verdict. Retention decay, content burn-rate vs production-rate,
economy entropy (inflation drift), and progression saturation predict trouble
months early. They cannot measure fun, social glue or creativity. Report both
the signal and its limit. See `references/unified-theory.md`.

## The canonical loop (use this for every task)

1. **Define value** — pick a base unit (often 1 minute of meaningful play, or
   1 unit of premium currency). Convert every resource with exact rational rates.
2. **Design additive-first** — every curve is `base + step * (level - 1)` with a
   cap. Only introduce multiplication with a stated reason and a hard cap.
3. **Validate** — check monotonicity, cap sanity, adjacent-level growth ratio,
   source/sink balance. Reject explosions before shipping.
4. **Ingest reality** — stream the project's config tables (up to 1,000,000
   xlsx/json/csv files); compare observed distributions to the design.
5. **Judge** — score the build with the 5-pillar health report; hand the AI the
   report + the diffs; let the agent decide what to change.
6. **Let the agent design** — the agent owns every number; it writes the configs,
   runs the validator, and explains each choice in terms of value/flow/pacing.

## Unit value & normalization (归一化)

Choose a base unit; define exact exchange rates; convert transitively. Never use
float for rates. `1/3` stays `1/3`.

```go
l, _ := normalize.NewLedger("金币", []normalize.Unit{
    {ID: "金币"}, {ID: "钻石"}, {ID: "体力"},
}, []normalize.Rate{
    {From: "钻石", To: "金币", Value: num.MustFromString("10")},
    {From: "体力", To: "金币", Value: num.MustFromString("2")},
})
gold, _ := l.Normalize("钻石", num.MustFromString("3")) // 30 金币, exact
```

CLI: `go run ./cmd/normalize spec.json`. Depth: `references/normalization.md`.

## Additive-first, not multiplicative (加法优先)

- **Default:** `value(n) = base + step*(n-1)`, capped. Linear, predictable, easy to
  rebalance, and the marginal cost of one more level is constant.
- **Why not multiplication:** compounding creates 数值膨胀 / power creep, boundary
  explosions, economy collapse via 复利, distorted probability expectations,
  runaway pay-to-win gaps, and debugging nightmares.
- **No global exponent, even if capped.** `base * 1.2^n` compounds long before the
  cap; the player-visible curve is still explosive. Use piecewise additive bands.
- **Multiplication is legitimate only when** it is a single controlled, capped
  multiplier, or percentage modifiers are **summed additively in one bucket**
  before one final multiply.
- Enforce it: `go run ./cmd/numgen -formula 'base*1.2^level' plan.json` fails with
  an explicit "global exponent is forbidden even with a cap" issue.
- Industry pitfall table + exact rules: `references/unified-theory.md`.

## Five pillars / eleven metrics (the health report)

| Pillar | Metrics |
|---|---|
| Retention | `retention_d1`, `retention_d7`, `retention_d30` |
| Economy | `economy_source_sink_ratio`, `economy_inflation` |
| Monetization | `monetization_pay_rate`, `monetization_arpu`, `monetization_ltv` |
| Progression | `progression_pace` |
| Content cadence | `content_cadence_days`, `content_debt_ratio` |

Fitness `= clamp(1 - |observed-ideal| / ((max-min)/2), 0, 1)`; status `critical`
outside `[min,max]`, `warn` below `0.6`, else `ok`; score is the weighted mean;
verdict `healthy >= 0.8`, `attention >= 0.6`, else `critical`. Full definitions:
`references/metrics.md`. Genre coverage/levers: `references/genres.md`.

CLI (values are strings in JSON, so precision survives — see `examples/gacha.json`):

```sh
go run ./cmd/opsreport examples/gacha.json     # score + verdict + diagnostics
go run ./cmd/econcheck examples/econ.json      # source/sink balance (~1)
```

`econcheck` sums `sources` and `sinks` maps of string values and reports the exact
ratio; >1.1 inflation, <0.9 scarcity, otherwise balanced. Specify every metric
key of the genre's levers in `Values`; a missing key is an error, not zero.

## Combat / growth / monetization numbers

Model all three as additive curves with value budgets and caps:

```go
plan := design.Plan{
    Growth:      design.AdditiveCurve("growth", "100", "25", "10000"),
    CombatPower: design.AdditiveCurve("power", "10", "3", "0"),     // 0 = no cap
    Damage:      design.AdditiveCurve("damage", "5", "2", "0"),
    HP:          design.AdditiveCurve("hp", "50", "12", "0"),
    Reward:      design.AdditiveCurve("reward", "20", "4", "0"),
    Price:       design.AdditiveCurve("price", "6", "1", "648"),
}
issues := plan.Validate(num.MustFromString("1.5")) // adjacent-level ratio cap
```

CLI: `go run ./cmd/numgen plan.json`. Rules for budgets, cap placement, pay-rate
discipline and why reward must never outgrow sink: `references/ai-design-workflow.md`.

## Streaming 1,000,000 config files & AI judgment

`ingest.Scan` walks roots with a bounded worker pool (never loads all files into
memory), parsing .xlsx (stdlib zip+xml), .json/.jsonl, .csv and .tsv, aggregating
exact per-column min/max/sum. Up to 1,000,000 files.

```sh
go run ./cmd/ingest -roots ./Config,./Tables -max 1000000 -workers 16 > stats.json
```

Give the agent: (1) the design plan, (2) `stats.json` observed distributions,
(3) the health report. Ask it to judge the current build and propose concrete
config edits, then re-run validation. It owns every number. Workflow + prompt:
`references/ai-design-workflow.md`.

## Flow (心流) analysis with codegraph233 — optional, slower, NOT tested here

codegraph233 (`neko233-com/codegraph233`) is a Rust AST GraphRAG MCP server for
game projects (code, prefab/scene/shader, config tables). To evaluate 心流, the
agent should auto-provision it at PROJECT level on demand:

```sh
# script auto-downloads a PINNED release into <project>/.codegraph233/bin/
scripts/codegraph-setup.ps1                    # or .sh on macOS/Linux
# invoke the project-local binary explicitly (it is NOT added to PATH):
.codegraph233/bin/codegraph233.exe init        # Windows
.codegraph233/bin/codegraph233 init && codegraph233 index && codegraph233 status
# then query via MCP (35 tools, auto-discovered): cg233_design_graph,
# cg233_config_tables, cg233_call_graph, cg233_asset_tree
```

MCP wiring, version pinning and config schema: `references/ai-design-workflow.md`.

Never block a numbers task on this; flow analysis takes much longer. Documented
in `references/ai-design-workflow.md`. This integration is deliberately NOT
verified in this repo's tests.

## Install (opencode / codex / dsh / claude — project first)

Copy this repo, then read the root `README.md`. One command installs the skill
into every applicable runtime's **project** directory; if no project is detected
it falls back to global. Use `-Scope all` to write both project and global:

```sh
scripts/install-skill.sh               # macOS/Linux (project)
scripts/install-skill.ps1              # Windows (project)
scripts/install-skill.ps1 -Scope global # global only
scripts/install-skill.ps1 -Scope all    # project AND global
```

Details and manual paths: `references/installation.md`.

## Common mistakes

| Mistake | Fix |
|---|---|
| Multiplicative stacking everywhere | Additive first; one capped multiplier; sum % in one bucket |
| float rates / rounding | Exact rational `num.Num`; string configs |
| No base unit | Normalize everything to one value unit before comparing |
| Missing metrics assumed zero | Error on missing values |
| Closed genre enum | Levers + open `genre.Register` |
| Hardcoded numbers in logic | Data-driven curves + validator; agent owns the numbers |
| Loading all config files into RAM | Bounded streaming `ingest.Scan` |
| Trusting numbers alone on lifespan | Report leading indicators + their limits |

## References

- `references/unified-theory.md` — 大道归一, lifespan, additive-first, pitfalls
- `references/normalization.md` — unit value, exchange rates
- `references/ai-design-workflow.md` — agent designs all numbers, 1M files, codegraph233
- `references/metrics.md` — eleven metrics
- `references/genres.md` — genre coverage
- `references/installation.md` — multi-runtime install
