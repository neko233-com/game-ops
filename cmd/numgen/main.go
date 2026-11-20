// Command numgen — thin launcher kept for `go run ./cmd/numgen` compatibility.
// Implementation lives in internal/commands/numgen (package numgencmd), shared with gameops.
package main

import numgencmd "github.com/neko233-com/game-ops/internal/commands/numgen"

func main() { numgencmd.Main() }
