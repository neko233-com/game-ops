// Command normalize — thin launcher kept for `go run ./cmd/normalize` compatibility.
// Implementation lives in internal/commands/normalize (package normalizecmd), shared with gameops.
package main

import normalizecmd "github.com/neko233-com/game-ops/internal/commands/normalize"

func main() { normalizecmd.Main() }
