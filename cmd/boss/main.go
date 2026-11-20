// Command boss — thin launcher kept for `go run ./cmd/boss` compatibility.
// Implementation lives in internal/commands/boss (package bosscmd), shared with gameops.
package main

import bosscmd "github.com/neko233-com/game-ops/internal/commands/boss"

func main() { bosscmd.Main() }
