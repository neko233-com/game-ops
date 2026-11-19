# Plan: Game Long-Term Operation Skill + Engine

1. Scaffold module, go.mod (go 1.27), .gitignore (test content ignored), design doc.
2. Package `num`: arbitrary-precision rational type. TDD: tests first (gitignored),
   watch RED, implement, watch GREEN.
3. Package `genre`: universal live-ops metrics, broad built-in registry for all
   market categories, open registration. TDD.
4. Package `ops`: genre-agnostic health evaluation engine. TDD.
5. Package `formula`: arbitrary numeric expression evaluator. TDD.
6. `cmd/opsreport`: JSON-config CLI; example config (tracked). Smoke test.
7. Author the skill (`skills/game-long-term-operation` + `.opencode/skills` copy)
   using writing-skills methodology (baseline scenario, skill, re-test).
8. Verify: `CGO_ENABLED=0 go build ./...`, `go vet ./...`, `go test ./...`,
   dependency check (stdlib only). Prove test content is gitignored via
   `git init` + `git check-ignore -v` + `git status --ignored`.
