// Command strategist — thin launcher kept for `go run ./cmd/strategist` compatibility.
// Implementation lives in internal/commands/strategist (package strategistcmd), shared with gameops.
package main

import strategistcmd "github.com/neko233-com/game-ops/internal/commands/strategist"

func main() { strategistcmd.Main() }
