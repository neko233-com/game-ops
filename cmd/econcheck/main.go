// Command econcheck — thin launcher kept for `go run ./cmd/econcheck` compatibility.
// Implementation lives in internal/commands/econcheck (package econcheckcmd), shared with gameops.
package main

import econcheckcmd "github.com/neko233-com/game-ops/internal/commands/econcheck"

func main() { econcheckcmd.Main() }
