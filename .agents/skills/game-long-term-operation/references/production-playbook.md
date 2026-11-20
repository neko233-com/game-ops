# Production playbook (数值主策划 / lead designer)

This reference is for the professional workflow: exact arithmetic, parallel
analysis at scale, time-to-cap modelling, and deterministic reproducible output.

## 1. Time-to-cap (常规玩家多久满级)

The question is never "what is the level-60 XP" but **how many calendar days a
specific cohort needs to reach it**. Model it exactly:

1. **Total XP** — `Σ(l=1..cap−1) [base + step·(l−1)]`, summed with `num.Num`.
2. **Daily XP** — allocate `MinutesPerDay` to activities in order, respecting
   each `DailyCapMinutes`: `Σ(minutes · XPPerMinute) · Efficiency`.
3. **Days** — `total / daily` as an exact rational; keep the fraction for audit,
   print the decimal for reading.
4. **Retention adjustment** — linearly interpolate the D1/D7/D30 anchors and
   extrapolate the D7→D30 slope to the cap day. `SurvivalAtCap` = the share of
   the cohort still present when they would reach cap. A low value means the
   cap is effectively unreachable for most players — move rewards earlier or
   lower the cap rather than celebrating the curve.
5. **Population percentiles** — P10/P25/P50/P75/P90 from archetype weights,
   sorted by days. P50 (not the mean) is the canonical "normal player" answer.

```sh
go run ./cmd/proanalysis -spec examples/proanalysis.json -workers 16 -repeat 20
```

`-repeat` multiplies the sensitivity grid (step variants × max-level variants)
to stress the parallel path. Output is a dense Markdown report: method, exact
per-archetype table, percentiles, full sensitivity grid, config outliers, and
recommendations with exact numbers. No boss narrative.

## 2. Parallel analysis at scale

- `progression.Batch(scenarios, workers)` uses a bounded worker pool and writes
  results into an index-addressed slice: deterministic order regardless of
  scheduling, no shared mutable state, no locks on result data.
- `ingest.Scan` uses a bounded channel (`workers*2`) and a worker pool; at most
  `workers` files are open at once. `MaxFiles` is capped at 1,000,000 by default
  and configurable via CLI; raise deliberately, never accidentally.
- `cmd/proanalysis` runs the scenario batch and the config scan **concurrently**
  (two goroutines joined with `sync.WaitGroup`, separate result variables).
- Tune `-workers` to the machine; correctness must not depend on the value.
  Verify with `-repeat` and two different worker counts that outputs match.
- Never aggregate via float maps; `Aggregate` uses exact `Min/Max/Sum` rationals.

## 3. Cross-file QA (what to flag automatically)

- Numeric columns with `max/min > 10` → inspect for unit errors or exponent
  formulas.
- Duplicate metric keys across config shards → define an ownership rule.
- A column whose adjacent-level ratio exceeds the configured `maxGrowthRatio`
  → multiplicative explosion; reject with `design.RejectGlobalExponent`.
- Missing values are errors, never zeros.
- Retention anchors inconsistent (D30 > D7 > D1 must hold) → reject the spec.

## 4. Presentation for designers (Vercel style)

`cmd/htmlreport` emits a self-contained Vercel-styled document: monochrome
palette (`#000`/`#fff`/`#eaeaea`), single accent `#0070f3`, Geist/Inter type
stack, `ui-monospace` for every number, hairline borders, no gradients or heavy
shadows. It contains exact fractions, the scoring formula, dense metric tables,
a source/sink diagram and an additive growth chart. Verify with:

```sh
go run ./cmd/htmlreport examples/report.json examples/dashboard.html
```

The docs site (`docs-site/`) uses **React Flow** for interactive value-flow and
system-causal diagrams, and **ECharts** for professional analytics (growth
comparison, time-to-cap percentiles, Sankey, sensitivity heatmap, config
outlier scatter). Charts must show legends, axes, tooltips and exact values —
they are instruments, not decoration. Never chart `float64` results as if exact;
compute the numbers with `num.Num` and only convert at the rendering boundary.

## 5. Production QA checklist

- [ ] `gofmt -l` empty, `go vet ./...` clean, `CGO_ENABLED=0 go build ./...` clean
- [ ] `go test -count=1 ./...` green with gitignored tests — `*_test.go` and
  fixtures are never committed; `examples/*.json` are public CLI samples, not
  test fixtures, and must contain no real project data
- [ ] same analysis with two different `-workers` values → identical results
- [ ] all config values parsed as strings; no `float64` in config path
- [ ] missing metric / unknown genre / zero divisor covered by tests
- [ ] docs site `npm run build` + `npm run check` + `tsc --noEmit` green
- [ ] generated HTML has zero external URLs/scripts (offline-safe)
- [ ] codegraph233 flow pass kept separate and explicitly unverified
