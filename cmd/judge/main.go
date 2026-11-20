// Command judge — thin launcher kept for `go run ./cmd/judge` compatibility.
// Implementation lives in internal/commands/judge (package judgecmd), shared with gameops.
package main

import judgecmd "github.com/neko233-com/game-ops/internal/commands/judge"

func main() { judgecmd.Main() }
