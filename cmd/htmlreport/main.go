// Command htmlreport generates a self-contained HTML dashboard for bosses and
// non-numerical readers from an exact game-ops analysis.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/neko233/game-ops/report"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "usage: htmlreport <spec.json> [output.html]")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read spec:", err)
		os.Exit(1)
	}
	var spec report.Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintln(os.Stderr, "parse spec:", err)
		os.Exit(1)
	}
	view, err := report.Build(spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, "build report:", err)
		os.Exit(1)
	}
	html, err := report.Render(view)
	if err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		os.Exit(1)
	}
	if len(os.Args) == 3 {
		if err := os.WriteFile(os.Args[2], html, 0644); err != nil {
			fmt.Fprintln(os.Stderr, "write html:", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", os.Args[2], len(html))
		return
	}
	os.Stdout.Write(html)
}
