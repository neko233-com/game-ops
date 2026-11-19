---
name: game-long-term-operation
description: Use when designing, auditing, balancing, or tuning long-term (live) game operations — 长线运营, retention, economy source/sink, inflation, monetization/LTV, progression pace, content cadence — for any game genre, or when numeric configs must support arbitrary-precision values (big.Rat, big.Int, fractions, huge numbers) and test content must never be pulled by collaborators. Provides the five-pillar metric taxonomy, a genre-agnostic lever model, exact-arithmetic rules, a health-scoring workflow, and a gitignore-based test-privacy recipe.
---

# Game Long-Term Operation (游戏长线运营)

## Overview

Long-term operation is a **control loop**, not a content calendar. The same loop
works for every genre because it runs on universal signals, not on genre-specific
features. Core principle:

> Describe a game as observed values for a fixed set of universal metrics; score
> them against a genre's target levers with exact arithmetic; act on the
> worst deviations; repeat on a fixed cadence.

This skill is supported by a pure-Go engine (`num`, `genre`, `ops`, `formula`,
`cmd/opsreport`). Use it directly when available; otherwise reproduce its model.

## The five pillars (eleven universal metrics)

| Pillar | Metrics | What it answers |
|---|---|---|
| Retention | `retention_d1`, `retention_d7`, `retention_d30` | Do players come back? |
| Economy | `economy_source_sink_ratio`, `economy_inflation` | Is currency balanced and non-inflationary? |
| Monetization | `monetization_pay_rate`, `monetization_arpu`, `monetization_ltv` | Is revenue healthy and sustainable? |
| Progression | `progression_pace` | Are players consuming content at the intended speed? |
| Content cadence | `content_cadence_days`, `content_debt_ratio` | Is new content arriving faster than it is burned? |

Full definitions and typical target bands: `references/metrics.md`.

## The genre-agnostic method

1. **Never hardcode a genre in the logic.** Logic reads metric keys only.
2. Describe each genre as a set of **levers**: one per relevant metric, with
   `Ideal`, `Min`, `Max` (the tolerance band) and `Weight` (importance). See
   `references/genres.md` for the built-in coverage and how levers differ.
3. Use the built-in registry when the genre is known; **register a new genre from
   data** when it is not. "All genres" means an open registry, never a closed enum.
4. An unknown genre is never a reason to invent new logic — add levers with the
   same eleven keys and register it.

## Arbitrary-precision rule

- Use an exact rational type (`math/big.Rat` in Go, `BigInt` numerator/denominator
  normalized by GCD elsewhere). **Do not roll your own** when a proven one exists;
  **do not use float64 for config math** — `0.1 + 0.2` must be exactly `3/10`.
- Store config values as **strings** (decimals, fractions like `1/3`, or scientific
  notation). Parse to exact numbers at the boundary, never the reverse.
- Convert to float **only** at the display/telemetry boundary.
- Integer powers and ratios stay exact; `big.Rat` has no magnitude ceiling.
- Keep numeric expressions data-driven: evaluate formulas over the exact type
  (`formula.Eval("arpu * pay_rate", vars)`), never with host-language floats.

## Validation workflow (the control loop)

1. **Describe**: pick the genre (or register it), list its levers.
2. **Observe**: build a snapshot — one exact value per lever key. A missing value
   is an error, not a zero; silent zeros corrupt the score.
3. **Score**: for each lever compute
   `fitness = clamp(1 - |observed - ideal| / ((max - min)/2), 0, 1)`.
   `Status` is `critical` outside `[min,max]`, `warn` below 0.6 fitness, else `ok`.
   Overall `score` is the **weighted mean** of fitness — all exact arithmetic.
4. **Judge**: `score >= 0.8` → `healthy`; `>= 0.6` → `attention`; else `critical`.
5. **Act**: fix the worst `critical`/`warn` levers first (largest ideal deviation
   weighted by importance), then re-snapshot on the genre's content cadence.

In Go:

```go
g := genre.MustGet("gacha") // or: genre.Register(myGenre); g := genre.MustGet(id)
values := map[string]num.Num{
    genre.MetricRetentionD1: num.MustFromString("0.44"), // string → exact
    // ... one entry per lever in g.Levers
}
report, err := ops.Evaluate(ops.Snapshot{GenreID: g.ID, Values: values})
// report.Score (exact num.Num), report.Verdict, report.Diagnostics
```

End-to-end CLI: `go run ./cmd/opsreport examples/gacha.json` (JSON values are
strings, so precision survives serialization).

## Keeping test content private

Test content must never be pulled by other developers. Prefer the simplest
mechanism that has **no external dependencies and no secrets**:

- Add to `.gitignore`: `*_test.go`, `/tests/`, `**/testdata/`, `*.test`,
  `coverage.out`.
- Keep all tests under `/tests/` as black-box tests of the public API (plus any
  `*_test.go` anywhere is ignored). Run them locally with `go test ./...`.
- Prove it: `git check-ignore -v tests/x_test.go` and `git status --ignored`.
- Avoid heavier schemes (private repos, submodules, CI tokens) unless policy
  requires them; they add secret-management burden and break contributor
  reproducibility for no gain here.

## Common mistakes

| Mistake | Why it fails | Do this instead |
|---|---|---|
| Hardcoding genre names/branches in logic | Breaks for the next genre; not "all genres" | Generic keys + levers + open registry |
| float64 money/rates | Rounding, overflow, `0.1+0.2 != 0.3` | Exact rational type, string configs |
| Rolling a custom BigRational | Bugs in normalize/GCD/serialization | Use `math/big.Rat` / proven BigInt |
| Treating missing metrics as 0 | Silently tanks the score | Error on missing lever keys |
| Unweighted or unnormalized scoring | One bad metric dominates or score is meaningless | Bounded fitness + weighted mean |
| Private test repo / CI secrets | Over-engineered, hurts reproducibility | `.gitignore` the test files |
| Skipping the re-snapshot cadence | A one-off audit is not "operation" | Re-run on the content cadence |

## References

- `references/metrics.md` — all eleven metrics, definitions, target bands, diagnostics.
- `references/genres.md` — built-in genre coverage and lever overrides.
