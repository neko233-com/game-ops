# Changelog

All notable changes to the game long-term-operation engine.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
versions follow [Semantic Versioning](https://semver.org/).

## [1.0.0] - 2026-11-20

### Added
- **Single production binary `gameops`**: all 15 capabilities in one statically-linked
  Go binary; 6 platforms (Windows/Linux/macOS × amd64/arm64); SHA256SUMS + version.json.
- **Boss decision center**: `boss brief|decide|monitor|duty|simulate|ask` — turns exact
  numbers into GO/HOLD/STOP verdicts, sign-off items with owner/deadline, and
  executable instructions. 26-job duty matrix (主策 12 + 运营 14). Vercel-style
  self-contained HTML cockpit.
- **Open catalog (支持所有游戏品类)**: `catalog derive|register|eval|strategy|export` —
  data-driven JSON specs, 14 tag-derivation rules, `GetAny`/`EvaluateAny`/`RunAny`
 宽松 API auto-derive any unseen genre; strict API semantics preserved.
- **Numerical engine**: exact rational `num.Num`, additive-first curves, global-exponent
  guard, unit normalization, gacha probability (binomial/geometric), 29 genre presets,
  streaming ingest (1M files), time-to-cap analysis, exact-fit health scoring.
- **Reports**: Vercel-style self-contained HTML (zh/en/ja), React Flow + ECharts docs site.
- **Multi-runtime skill install**: opencode / claude / codex / dsh (project-first).
- LICENSE (MIT), CI workflows (build/test/release), production release scripts.

### Notes
- `go test -race` requires CGO + a C compiler; not run in the release environment
  (no gcc). Determinism verified by repeated parallel-run hash comparison instead.
- Strict API (`Register`/`Get`/`Evaluate`/`Run`) keeps original error semantics;
  the new `*Any` APIs power the "support everything" path.

[1.0.0]: https://github.com/neko233-com/game-ops/releases/tag/v1.0.0
