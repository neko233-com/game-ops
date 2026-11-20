// Command proanalysis — thin launcher kept for `go run ./cmd/proanalysis` compatibility.
// Implementation lives in internal/commands/proanalysis (package proanalysiscmd), shared with gameops.
package main

import proanalysiscmd "github.com/neko233-com/game-ops/internal/commands/proanalysis"

func main() { proanalysiscmd.Main() }
