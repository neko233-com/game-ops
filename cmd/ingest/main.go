// Command ingest scans config tables and prints exact per-column statistics.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/neko233-com/game-ops/ingest"
)

func main() {
	roots := flag.String("roots", ".", "comma-separated directories to scan")
	maxFiles := flag.Int64("max", 1000000, "maximum number of files (ceiling 1,000,000)")
	workers := flag.Int("workers", 8, "number of parallel workers")
	flag.Parse()

	stats, err := ingest.Scan(context.Background(), ingest.Options{
		Roots: strings.Split(*roots, ","), MaxFiles: *maxFiles, Workers: *workers,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(stats); err != nil {
		fmt.Fprintln(os.Stderr, "encode:", err)
		os.Exit(1)
	}
}
