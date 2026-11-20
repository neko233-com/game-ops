// Command formula — thin launcher kept for `go run ./cmd/formula` compatibility.
// Implementation lives in internal/commands/formula (package formulacmd), shared with gameops.
package main

import formulacmd "github.com/neko233-com/game-ops/internal/commands/formula"

func main() { formulacmd.Main() }
