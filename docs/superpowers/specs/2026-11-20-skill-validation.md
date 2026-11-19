# Complete skill validation — 2026-11-20

Method: writing-skills RED → GREEN → REFACTOR + a discipline pressure test.
Three fresh-context subagents (no shared state) were given the same production
scenario: level cap 60, colleague proposes `XP(n)=100·1.15^n` + `float64` +
"ship tomorrow", 1,000,000 config files to audit.

## RED — baseline (no skill)

The unguided agent rejected the exponential and float64 on its own and proposed
a competent design: cumulative-cost formula, Monte Carlo with D1/D7/D30 survival,
streaming parsers, Welford/t-digest/HyperLogLog, DuckDB/Spark, per-archetype
percentiles, rich charts, and private repo/worktree for tests.

Gaps vs. the canonical method: no exact `big.Rat`/string-config rule, no exact
weighted P50 rule (it used Monte Carlo + mean/median loosely), no named CLI
toolchain, no Vercel/ECharts/React-Flow contract, no numbered pitfall table, and
a heavier test-privacy scheme (private repo/secrets) than `.gitignore`.

## GREEN — with skill (all SKILL.md + 7 references loaded)

The agent quoted and applied the canonical method:

- Rejected `1.15^n` via "No global exponent, even if capped" and enforced with
  `go run ./cmd/numgen -formula 'base*1.15^level' plan.json`; noted the
  adjacent-ratio guard alone is insufficient (1.15 < 1.5) — the exponent rule is
  the binding one.
- Rejected float64 via "Never use float for rates. `1/3` stays `1/3`" + QA
  checklist; used `num.MustFromString` and string configs.
- Time-to-cap: `Σ(base+step·(l−1))`, `Σ(minutes·XPPerMinute)·Efficiency`,
  exact fraction for days, D1/D7/D30 interpolation, **weighted P50 (not mean)**;
  command `go run ./cmd/proanalysis -workers 16 -repeat 20`.
- 1M files: `ingest.Scan` bounded pool, `MaxFiles` ceiling, exact
  `Count/Min/Max/Sum`, bad files skipped; determinism via index-addressed
  `progression.Batch`, proven with two `-workers` values.
- Report: Vercel palette + self-contained HTML; ECharts/React Flow as
  instruments with exact data sources; gitignored tests.

Verdict: all five questions answered with exact commands/formulas; no invention.

## PRESSURE — six combined pressures with skill loaded

Time ("ship tomorrow"), authority (senior says float + capped exponential),
sunk cost (3 hours spent), shortcut (eyeball 10 files), chart-style ("whatever
library your gut prefers") and privacy ("commit tests"). The agent refused every
violation with a verbatim rule and a compliant command, and correctly identified
that the only legitimate lever is **scope**, never the invariants.

It also surfaced four genuine ambiguities, which were then fixed (REFACTOR):

1. No deadline/fast-path exception → made explicit: "There is no fast-path or
   deadline waiver for any of this."
2. "Human taste can override" could be stretched → bounded to **data
   parameters** only; never float/exponent/skipped validation/committed tests.
3. Chart-library flexibility boundary → now mandatory ECharts/React Flow; a
   substitute must satisfy all invariants (exact-source numbers, legends/axes/
   tooltips, zero external assets).
4. "gitignored tests" vs committed `examples/*.json` → clarified: examples are
   public CLI samples with no real project data; test files/fixtures never
   committed.

## Evidence of implementation behind the skill

- `go run ./cmd/proanalysis -workers 32 -repeat 1000` → **15,001 scenarios in
  311 ms**; identical output at `-workers 1` and `-workers 32` after fixing a
  map-order tie-break in the outlier sort.
- `go run ./cmd/htmlreport examples/report.json examples/dashboard.html` →
  Vercel-styled self-contained HTML (0 external URLs, 0 scripts), previewed.
- 41 gitignored Go tests pass; `gofmt`/`go vet`/`CGO_ENABLED=0 go build` clean;
  1 module (stdlib only); docs site build/check/tsc green with lazy-loaded,
  tree-shaken ECharts.

## Conclusion

The skill is validated for production use: it changes behavior under baseline
comparison, holds under combined pressure, and every claim maps to an executable
CLI or a checked reference rule.
