# Design addendum: unified numerical theory, AI design, docs site, multi-runtime

Date: 2026-11-19
Status: Approved-by-directive (human asked for expansion, no Q&A gate)

## Added intent

Extend the long-term-operation skill from a health-scoring engine into a complete
numerical-design system:

1. 单位价值 / 归一化 — compare all resources in one base unit with exact rates.
2. 大道归一 — teach that the core of 数值 is 价值的定义、流动、节奏; state honestly
   that lifespan is a leading indicator, not a verdict.
3. 加法优先 — additive-first curves; document industry pitfalls of multiplication.
4. AI designs all numbers — agent owns combat/growth/monetization values end to end.
5. Million-file config analysis — bounded streaming ingest of xlsx/json/csv/tsv.
6. codegraph233 flow analysis — optional, project-level auto-provision, NOT tested here.
7. Companion docs site — Vite+ (vite-plus) + React + TypeScript, all concepts.
8. Multi-runtime install — opencode / codex / dsh / claude, project-first.

## Decisions (my calls)

- `num` stays the exact base; `normalize` builds a rate graph on top of it (BFS,
  transitive, exact) — no float anywhere.
- `ingest` is stdlib-only: xlsx via `archive/zip` + `encoding/xml`; JSON via
  streaming `Decoder`; CSV/TSV via `encoding/csv`. Bounded worker pool = bounded
  memory; `MaxFiles` caps at 1,000,000.
- `design` uses additive curves `base + step*(n-1)` with optional cap and an
  adjacent-level growth-ratio guard; multiplication only as a single capped bucket.
- Skill references are progressive: `SKILL.md` (<500 lines) holds the method;
  six focused `references/*.md` hold the depth.
- Canonical skill at `skills/game-long-term-operation/`; installers copy it to
  each runtime's project dir first, global second. The in-repo `.opencode/` copy
  is kept synced so the harness discovers it immediately.
- codegraph233 is provisioned into `<project>/.codegraph233/bin/` on demand;
  explicitly excluded from this repo's verification.
- Docs site lives at `docs-site/` inside the same repo (so "clone → README → use").

## Professional pivot (lead numerical designer)

The audience was corrected from "bosses" to **数值主策划**. Added:
- `progression`: exact time-to-cap with archetypes, daily time budgets, D1/D7/D30
  survival adjustment, weighted P10–P90, and deterministic parallel `Batch`.
- `cmd/proanalysis`: concurrent scenario batch + config scan, `-workers`,
  `-repeat` stress grid, dense Markdown report with exact fractions and
  sensitivity table. Fixed a map-order tie-break so results are identical at
  worker counts 1 and 32 (verified).
- HTML report restyled to **Vercel design language** (monochrome, hairline
  borders, Geist/Inter, monospace numbers, `#0070f3` accent), with exact
  fractions and formula text instead of boss narrative.
- ECharts + React Flow in the docs site as designer instruments (not decoration).
- `references/production-playbook.md` + full skill test suite (baseline,
  with-skill, pressure) to prove the skill holds under production conditions.

## Final additions: executive communication

- `report` + `cmd/htmlreport`: self-contained HTML dashboard (inline CSS/SVG,
  offline/email-safe) translating exact analysis into boss-friendly visuals:
  health gauge, five pillar cards, traffic-light table, money-flow diagram,
  additive-vs-multiplicative chart, three decisions list. `examples/dashboard.html`
  is a committed generated artifact.
- `cmd/judge`: assembles an AI-ready packet (exact top deviations, config
  outliers from `ingest` stats, copyable prompt) so "AI judges the build" is
  executable, not just advice.
- `docs-site` React Flow pages: interactive 价值流动图 and 系统关系图 with an
  additive-vs-multiplicative toggle, for non-numerical readers.

## Verification plan

- Go: `gofmt`, `go vet ./...`, `CGO_ENABLED=0 go build ./...`, `go test -count=1 ./...`,
  smoke all four CLIs, gitignore proof.
- Docs: `npm install`, `npm run build`, `npm run check` in `docs-site/`.
- Skill: harness lists `game-long-term-operation`; installer tested in a temp
  project with all four project targets; copies identical.
- codegraph233: NOT tested (as specified).
