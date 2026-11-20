// Command curve — thin launcher kept for `go run ./cmd/curve` compatibility.
// Implementation lives in internal/commands/curve (package curvecmd), shared with gameops.
package main

import curvecmd "github.com/neko233-com/game-ops/internal/commands/curve"

func main() { curvecmd.Main() }
