// Command numcalc — thin launcher kept for `go run ./cmd/numcalc` compatibility.
// Implementation lives in internal/commands/numcalc (package numcalccmd), shared with gameops.
package main

import numcalccmd "github.com/neko233-com/game-ops/internal/commands/numcalc"

func main() { numcalccmd.Main() }
