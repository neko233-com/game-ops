# game-ops — Game Long-Term Operation Engine & Skill

A pure-Go (Go 1.27, **no cgo**, standard library only) engine plus an agent skill
for designing and validating **long-term game operations** (游戏长线运营).

## Design goals

| Requirement | How it is met |
|---|---|
| Game long-term operation skill | `skills/game-long-term-operation/` (and `.opencode/skills/` copy) |
| Adapt to all game genres | Genre-agnostic engine; broad built-in registry + open `genre.Register` |
| Support arbitrary numeric values | `num.Num` backed by `math/big.Rat` (exact, unbounded) |
| Test content never pulled by others | `.gitignore` ignores `*_test.go`, `/tests/`, `**/testdata/` |
| Go 1.27, no cgo | `go 1.27` in `go.mod`, stdlib only, `CGO_ENABLED=0` build verified |

## Packages

- `num` — exact arbitrary-precision rational arithmetic (`big.Rat`). `0.1+0.2 == 3/10`,
  values of any magnitude, fractions, decimals, scientific notation, JSON round-trip.
- `genre` — universal live-ops metrics (retention, economy, monetization,
  progression, content cadence) and a built-in registry covering all major market
  categories. Custom genres can be registered from data at runtime.
- `ops` — `Evaluate(Snapshot) -> Report`: weighted, genre-parameterized health
  scoring with exact arithmetic and ok/warn/critical diagnostics.
- `formula` — evaluator for arbitrary numeric config expressions (`arpu * 0.04`,
  `2^3^2`, `1/3 + 2/3`) returning `num.Num`.
- `cmd/opsreport` — CLI: `go run ./cmd/opsreport examples/gacha.json`.

## Usage

```go
g := genre.MustGet("gacha")
values := map[string]num.Num{
    genre.MetricRetentionD1: num.MustFromString("0.44"),
    // ... one entry per lever in g.Levers
}
report, err := ops.Evaluate(ops.Snapshot{GenreID: g.ID, Values: values})
fmt.Println(report.Score.DecimalString(4), report.Verdict)
```

Config values are supplied as strings so arbitrary precision survives JSON.

## Test content

All tests live outside the tracked source set: `tests/*_test.go` is black-box
tests of the public API, and `.gitignore` excludes every `*_test.go`, `/tests/`
and `testdata/` directory. Run them locally with `go test ./...`; they will
never be committed or pushed.

## Verification

```powershell
$env:CGO_ENABLED = "0"
go build ./...
go vet ./...
go test ./...
```
