# Design: Game Long-Term Operation (游戏长线运营) Skill + Engine

Date: 2026-11-19
Status: Approved-by-directive (human opted out of Q&A via "start now, don't ask")

## Intent / Why

Build a reusable **skill** plus a **pure-Go engine** for designing and validating
long-term game operations ("长线运营") that:

1. works for **every game genre** on the market — no closed hardcoded set;
2. handles **arbitrary numeric values** — no float64/int64 ceiling, exact math;
3. keeps **all test content gitignored** so collaborators never pull it;
4. targets **Go 1.27, no cgo** — standard library only, `CGO_ENABLED=0`.

The deliverable is two-fold: (a) an agent skill (`skills/game-long-term-operation`)
teaching the universal method, and (b) the Go engine the skill references, proven
by gitignored tests.

## Goals

- A `num` type with exact arbitrary-precision arithmetic (backed by `math/big.Rat`).
- A genre system with a broad built-in registry AND an open registry so ANY genre
  (including future/unknown ones) can be registered from data.
- A genre-agnostic operation engine producing a health report over universal
  live-ops metrics, parameterized per genre.
- A formula evaluator for arbitrary numeric config expressions.
- A CLI that reads a JSON config (values as strings → arbitrary precision) and
  emits a report, proving end-to-end use.
- Gitignored black-box tests that prove every claim.

## Non-goals

- No network/perf analytics ingestion, no telemetry backend, no game client code.
- No external dependencies (keeps "no cgo" and reproducible builds trivial).
- No UI.

## Constraints

- Go 1.27.0, `go 1.27` in go.mod, `CGO_ENABLED=0`, stdlib only.
- All `*_test.go`, `/tests/`, and `**/testdata/` are gitignored.

## Approaches considered

1. **float64-based math** — rejected: overflows/precision loss, violates
   "任意数值".
2. **big.Float-based** — rejected as base: transcendental ops exist but equality
   and determinism are awkward; rational base is exact and simpler to validate.
3. **big.Rat-based (chosen)** — exact arbitrary-precision rationals; integer
   powers exact; division exact; decimals/fractions/scientific parsed from
   strings. Deterministic, pure Go, no cgo. Best fit for "支持任意数值".

For genre coverage: a **closed enum** was rejected (can't cover "all genres");
chosen is a **broad built-in registry + open registration** so coverage is
effectively unbounded and data-driven.

## Design decisions (my calls)

- Module `github.com/neko233/game-ops`, packages: `num`, `genre`, `ops`,
  `formula`, `cmd/opsreport`.
- Health score = weighted mean of per-metric fitness in [0,1], all computed with
  `num.Num`. Fitness = clamp(1 - |observed-ideal| / tolerance, 0, 1), where
  tolerance is derived from each lever's configured min/max band.
- Canonical skill lives at `skills/game-long-term-operation/`; a discoverable
  copy lives at `.opencode/skills/game-long-term-operation/`.
- Tests live in gitignored `/tests/` as black-box tests of the public API, plus
  any `*_test.go` anywhere is ignored. A local `git init` + commit proves test
  content is not tracked.

## What's left to the builder

Nothing requiring human input; implementation proceeds via TDD and final
verification (`go test ./...`, `go vet`, `go build` with `CGO_ENABLED=0`).
