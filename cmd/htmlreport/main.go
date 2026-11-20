// Command htmlreport — thin launcher kept for `go run ./cmd/htmlreport` compatibility.
// Implementation lives in internal/commands/htmlreport (package htmlreportcmd), shared with gameops.
package main

import htmlreportcmd "github.com/neko233-com/game-ops/internal/commands/htmlreport"

func main() { htmlreportcmd.Main() }
