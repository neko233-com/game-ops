---
name: game-long-term-operation
description: Use when designing, auditing, balancing or monetizing long-term game operations (游戏长线运营) and ALL numerical design (数值) — unit value & normalization (归一化/单位价值), additive-first formulas (加法优先), combat/growth/monetization numbers, an agent designing every number, streaming analysis of 1,000,000 Excel/JSON/CSV config files, judging a build's numerical health, predicting game lifespan from numbers, and optional codegraph233 flow (心流) analysis. Genre-agnostic, exact math, project-first install for opencode/codex/dsh/claude.
---

# Game Long-Term Operation & Numerical Design (游戏长线运营 / 数值)

## Production binary — single `gameops`

All 15 capabilities ship in ONE statically-linked Go binary. Build with
`scripts/build-release.sh` (or `.ps1`), 6 platforms, SHA256SUMS + version.json.
Every `go run ./cmd/<name>` example in this SKILL also works as
`gameops <name>` (e.g. `gameops opsreport examples/gacha.json`,
`gameops boss brief gacha`, `gameops numcalc '1/3 + 2/7'`).
`gameops version` prints version/commit/build-date/Go. Unknown command exit 1,
panic exit 70. See `RELEASE.md`.

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
# --- 长线运营健康 ---
go run ./cmd/opsreport examples/gacha.json           # score + verdict + diagnostics
go run ./cmd/opsreport examples/gacha-warn.json      # attention 级（示例）
go run ./cmd/econcheck examples/econ.json            # source/sink balance (~1)

# --- 数值生成与校验 ---
go run ./cmd/numgen examples/plan.json               # 加法优先曲线
go run ./cmd/numgen -formula 'base*1.2^level' examples/plan.json  # 指数守卫 → exit 1
go run ./cmd/curve examples/piecewise.json           # 多段加法曲线（见 references/piecewise.md）
go run ./cmd/curve example                           # 打印推荐默认值

# --- 精确计算 ---
go run ./cmd/formula '2^3^2'                         # 表达式求值 → 512
go run ./cmd/formula -expr 'a+b*c' -vars 'a=1,b=2,c=3'
go run ./cmd/numcalc '1/3 + 2/7'                     # → 13/21
go run ./cmd/numcalc comb 10 3                        # C(10,3) = 120
go run ./cmd/numcalc binom 10 3 0.06                  # 精确概率 C(10,3)·0.06³·0.94⁷
go run ./cmd/numcalc geom 80 0.006                    # 第 k 次才首次成功的概率（保底 SSR）
go run ./cmd/numcalc mean 10 20 30                    # mean / median / variance / pXX

# --- Query & 行业品类 ---
go run ./cmd/genre list                              # 29 内置品类
go run ./cmd/genre show gacha                        # gacha 的 11 项 lever 理想值
go run ./cmd/genre metrics                           # 11 个通用指标定义

# --- 开放品类目录（支持所有品类） ---
go run ./cmd/catalog list
go run ./cmd/catalog derive cozy-farming-sim -name 治愈农场 -tags cozy,farming -save g.json
go run ./cmd/catalog register g.json
go run ./cmd/catalog eval cozy-farming-sim
go run ./cmd/catalog strategy cozy-farming-sim -region CN
go run ./cmd/catalog export -out catalog.json

# --- 版本判断 & 摄取 ---
go run ./cmd/judge -values examples/gacha.json        # 给 AI 的版本判断包
go run ./cmd/ingest -roots ./Config -max 1000000 -workers 16 > stats.json
go run ./cmd/proanalysis -spec examples/proanalysis.json -workers 16 -repeat 20
go run ./cmd/htmlreport examples/report.json dashboard.html
```

`econcheck` sums `sources` and `sinks` maps of string values and reports the exact
ratio; >1.1 inflation, <0.9 scarcity, otherwise balanced. Specify every metric
key of the genre's levers in `Values`; a missing key is an error, not zero.

## Multi-segment piecewise additive (多段加法)

When pure `base + step·(n-1)` looks too flat at high levels and a capped
exponential is forbidden, the ONLY correct design is **piecewise additive
curves** — each range `[start, end]` has its own `base + step·(n-start)` with
an optional global hard cap. Every boundary is explicit design intent, not
hidden compounding. Full rules and why it replaces exponential:
`references/piecewise.md`.

```go
pw := design.AdditivePiecewise("xp", "10000",
    design.Piece(1, 10, "100", "10"),
    design.Piece(11, 30, "200", "15"),
    design.Piece(31, 60, "500", "20"),
)
```

## Probability & gacha math（抽卡概率 / 组合数学）

Any random system — gacha, drop rate, crit, skill proc — must use **exact
rational** probability (`float64` drifts after repeated products). The engine
provides `num.BinomialP`, `num.GeometricP`, `num.Comb`, `num.Perm`,
`num.Factorial`, and `numcalc` CLI. Full formulas and examples:
`references/probability.md`.

```sh
# 80 抽保底 SSR（掉率 0.6%）恰好第 80 抽时才首次命中的概率
go run ./cmd/numcalc geom 80 0.006
# 10 次独立 3% 暴击至少命中 2 次
go run ./cmd/numcalc binom 10 0 0.03
go run ./cmd/numcalc binom 10 1 0.03
# 期望值 E = Σ k·geom(k,p), 用 numcalc 累加
```

## Boss decision center — 代替主策+运营，给老板用（ conclusions first ）

When the user is a decision-maker (not a numerical specialist), do NOT dump
metrics/formulas. Use the boss layer: verdict → sign-off items → instructions.

```sh
go run ./cmd/boss brief <genre> [-values f.json]       # GO/HOLD/STOP + top decisions
go run ./cmd/boss decide <genre> [-values f.json] -html out.html  # full packet + cockpit
go run ./cmd/boss monitor <genre> [-values f.json]     # traffic-light ops dashboard
go run ./cmd/boss duty [-role 主策划|运营]               # 26-job duty matrix + standards
go run ./cmd/boss simulate <genre> [-values f.json]    # now / fixed / ignored scenarios
go run ./cmd/boss ask "gacha 这个版本能上吗"             # natural-language routing
```

Go API: `boss.Analyze(genreID, values, flags) -> Brief{Headline, Call, Findings,
Decisions, Actions, Duties, Limits}`, `boss.DutyMatrix` (26 jobs), `boss.Translate`
(key → plain language + 价值/流动/节奏 family). Unknown genres auto-derived — any
category works. Findings are exact numbers translated to business language;
each decision carries owner/deadline/options/if-not; limits state what numbers
cannot judge (fun, social, creativity). Never present the boss with a raw metric
table — always lead with the one-line verdict.

## Open catalog — support EVERY genre (支持所有品类)

The engine is **genre-agnostic by construction**: 11 universal metrics + a data-driven
catalog. A category does NOT need a built-in preset or a generated Strategy to be
analysable — unknown IDs are auto-derived from their name/tags (14 tag rules over
the universal levers), registered at runtime, and dispatched to a universal
fallback strategy. JSON specs let you override any lever without codegen or rebuild.

```sh
go run ./cmd/catalog list                                   # every registered genre
go run ./cmd/catalog derive cozy-farming-sim -name 治愈农场 -tags cozy,farming -save g.json
go run ./cmd/catalog register g.json                        # JSON Spec[] or single Spec
go run ./cmd/catalog eval cozy-farming-sim                  # health report, any ID
go run ./cmd/catalog strategy cozy-farming-sim -region CN   # universal fallback strategy
go run ./cmd/catalog export -out catalog.json               # round-trippable full catalog
```

Go API:
- `genre.Derive(id, name, tags) Genre` — build levers from tags, no registration.
- `genre.GetAny(id) Genre` — get-or-derive-and-register (never fails).
- `genre.Upsert / LoadFile / LoadDir / Export / ValidateSpec` — data-driven catalog.
- `ops.EvaluateAny(Snapshot)` — score ANY category; partial snapshots OK.
- `factory.GetAny / RunAny` — registered strategy, else universal strategy.
- `genre.Register` (strict, duplicate errors) and `ops.Evaluate` (strict, unknown
  errors) are kept unchanged for backward compatibility.

Rules stayed intact: 11 metrics, exact `num.Num`, additive-first, open `genre.Register`.
Only the **entry points** widened — nothing existing broke. `catalog export` output
round-trips via `catalog register`. Each genre spec is pure JSON with string numbers.

## Universal genre levers（29 built-in + open Register）

The engine ships with 29 genre presets covering gacha, RPG, MMO, MOBA, FPS,
SLG, casual, match-3, tower-defense, visual-novel, rhythm, and more. Each
genre customizes ideal values, tolerance bands and weights for the 11
universal metrics. Open `genre.Register` lets you add ANY genre including
future ones — not a closed enum.

```sh
go run ./cmd/genre list                    # 29 IDs
go run ./cmd/genre show slg                # slg 的完整 11 项 lever 表
go run ./cmd/genre metrics                 # 11 个通用指标定义
```

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
config edits, then re-run validation. It owns every number. Assemble the packet
with `go run ./cmd/judge -values examples/gacha.json -base 金币` (exact top
deviations, config outliers, copyable AI prompt). Workflow + prompt:
`references/ai-design-workflow.md`.

## Time-to-cap & parallel professional analysis

For the **lead numerical designer**, answer "how long until a normal player hits
max level" with exact arithmetic and player cohorts:

- total XP = `Σ base+step·(l−1)`; daily XP = time-budgeted activity XP × efficiency;
  days = exact fraction; **P50 of weighted archetypes is the normal player**.
- adjust with D1/D7/D30 survival: a low `SurvivalAtCap` means the cap is
  effectively unreachable, not a success.
- run hundreds/thousands of scenarios in parallel, deterministically, and scan
  config tables concurrently.

```sh
go run ./cmd/proanalysis -spec examples/proanalysis.json -workers 16 -repeat 20
```

Full production method, QA checklist and chart conventions:
`references/production-playbook.md`.

## Vercel-style professional report & charts

`cmd/htmlreport` emits a **self-contained Vercel-styled** HTML document for
designers: monochrome (`#000`/`#fff`/`#eaeaea`), single accent `#0070f3`,
Geist/Inter stack, monospace numbers, hairline borders, no gradients. It shows
exact fractions, the fitness formula, dense metric tables, source/sink flow and
the additive growth chart. Zero external URLs/scripts — offline/email safe.

```sh
go run ./cmd/htmlreport examples/report.json report.html
```

The docs site adds **React Flow** (interactive value-flow + system-causal
diagrams) and **ECharts** (growth comparison, time-to-cap percentiles, Sankey,
sensitivity heatmap, config outlier scatter). Charts are instruments: legends,
axes, tooltips and exact values; compute with `num.Num`, convert only at the
rendering boundary. Do not substitute an arbitrary chart library on preference:
use the provided ones; a different library is acceptable only if it satisfies
every invariant above (exact-source numbers, legends/axes/tooltips, zero
external assets). There is no fast-path or deadline waiver for any of this.

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
| Judging "normal player" by the mean | Weighted P50 of archetypes; report survival at cap |
| Charting float results as if exact | Compute with `num.Num`; convert only at render |
| Non-deterministic parallel results | Index-addressed result slice, fixed worker semantics |
| Trusting numbers alone on lifespan | Report leading indicators + their limits |

## Language (中文优先 / multi-language)

Chinese is the authoritative primary language and the default everywhere —
even when the environment locale is not Chinese. English and Japanese are
supported alternatives. Switch reports with `-lang zh|en|ja` on `cmd/htmlreport`
and `cmd/proanalysis`; the docs site has a `中文 / EN / 日本語` switcher that
defaults to Chinese and persists the choice. Numeric values and metric keys are
never translated; only labels and prose are. Details: `references/i18n.md`.

## References

- `references/boss-decision-center.md` — 老板决策中枢：六命令、Brief 结构、翻译表、26 项工作事务
- `references/open-catalog.md` — 支持所有品类：派生规则、JSON spec、宽松/严格 API、通用兜底
- `references/i18n.md` — Chinese-primary, CLI `-lang`, docs-site switcher
- `references/production-playbook.md` — time-to-cap, parallel scale, QA, Vercel/ECharts
- `references/unified-theory.md` — 大道归一, lifespan, additive-first, pitfalls
- `references/normalization.md` — unit value, exchange rates
- `references/ai-design-workflow.md` — agent designs all numbers, 1M files, codegraph233
- `references/metrics.md` — eleven metrics
- `references/genres.md` — genre coverage
- `references/installation.md` — multi-runtime install
- `references/piecewise.md` — multi-segment additive curves (only legit alternative to exponential)
- `references/probability.md` — gacha/drop-rate math, binomial/geometric, exact rational
