// Command ingest — thin launcher kept for `go run ./cmd/ingest` compatibility.
// Implementation lives in internal/commands/ingest (package ingestcmd), shared with gameops.
package main

import ingestcmd "github.com/neko233-com/game-ops/internal/commands/ingest"

func main() { ingestcmd.Main() }
