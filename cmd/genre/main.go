// Command genre — thin launcher kept for `go run ./cmd/genre` compatibility.
// Implementation lives in internal/commands/genre (package genrecmd), shared with gameops.
package main

import genrecmd "github.com/neko233-com/game-ops/internal/commands/genre"

func main() { genrecmd.Main() }
