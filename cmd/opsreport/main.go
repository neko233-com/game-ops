// Command opsreport — thin launcher kept for `go run ./cmd/opsreport` compatibility.
// Implementation lives in internal/commands/opsreport (package opsreportcmd), shared with gameops.
package main

import opsreportcmd "github.com/neko233-com/game-ops/internal/commands/opsreport"

func main() { opsreportcmd.Main() }
