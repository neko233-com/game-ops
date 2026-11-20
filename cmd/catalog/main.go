// Command catalog — thin launcher kept for `go run ./cmd/catalog` compatibility.
// Implementation lives in internal/commands/catalog (package catalogcmd), shared with gameops.
package main

import catalogcmd "github.com/neko233-com/game-ops/internal/commands/catalog"

func main() { catalogcmd.Main() }
